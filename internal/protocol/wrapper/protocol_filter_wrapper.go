package protocolwrapper

import (
	"context"
	"strings"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/filter"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
)

const (
	// FILTER is protocol key.
	FILTER = "filter"
)

func init() {
	extension.SetProtocol(FILTER, GetProtocol)
}

// ProtocolFilterWrapper
// protocol in url decide who ProtocolFilterWrapper.protocol is
type ProtocolFilterWrapper struct {
	protocol protocol.Protocol
}

// Export service for remote invocation
func (pfw *ProtocolFilterWrapper) Export(invoker protocol.Invoker) protocol.Exporter {
	if pfw.protocol == nil {
		pfw.protocol = extension.GetProtocol(invoker.GetURL().Protocol)
	}
	invoker = BuildInvokerChain(invoker, constant.ServiceFilterKey)
	return pfw.protocol.Export(invoker)
}

// Refer a remote service
func (pfw *ProtocolFilterWrapper) Refer(url *common.URL) protocol.Invoker {
	if pfw.protocol == nil {
		pfw.protocol = extension.GetProtocol(url.Protocol)
	}
	invoker := pfw.protocol.Refer(url)
	if invoker == nil {
		return nil
	}
	return BuildInvokerChain(invoker, constant.ReferenceFilterKey)
}

// Destroy will destroy all invoker and exporter.
func (pfw *ProtocolFilterWrapper) Destroy() {
	pfw.protocol.Destroy()
}

func BuildInvokerChain(invoker protocol.Invoker, key string) protocol.Invoker {
	filterName := invoker.GetURL().GetParam(key, "")

	if filterName == "" {
		return invoker
	}
	filterNames := strings.Split(filterName, ",")

	// The order of filters is from left to right, so loading from right to left
	next := invoker
	for i := len(filterNames) - 1; i >= 0; i-- {
		flt, _ := extension.GetFilter(strings.TrimSpace(filterNames[i]))
		fi := &FilterInvoker{next: next, invoker: invoker, filter: flt}
		next = fi
	}

	if key == constant.ServiceFilterKey {
		logger.Debugf("[BuildInvokerChain] The provider invocation link is %s, invoker: %s",
			strings.Join(append(filterNames, "proxyInvoker"), " -> "), invoker)
	} else if key == constant.ReferenceFilterKey {
		logger.Debugf("[BuildInvokerChain] The consumer filters are %s, invoker: %s",
			strings.Join(append(filterNames, "proxyInvoker"), " -> "), invoker)
	}
	return next
}

// nolint
func GetProtocol() protocol.Protocol {
	return &ProtocolFilterWrapper{}
}

// FilterInvoker defines invoker and filter
type FilterInvoker struct {
	next    protocol.Invoker
	invoker protocol.Invoker
	filter  filter.Filter
}

// GetURL is used to get url from FilterInvoker
func (fi *FilterInvoker) GetURL() *common.URL {
	return fi.invoker.GetURL()
}

// IsAvailable is used to get available status
func (fi *FilterInvoker) IsAvailable() bool {
	return fi.invoker.IsAvailable()
}

// Invoke is used to call service method by invocation
func (fi *FilterInvoker) Invoke(ctx context.Context, invocation protocol.Invocation) protocol.Result {
	result := fi.filter.Invoke(ctx, fi.next, invocation)
	return fi.filter.OnResponse(ctx, result, fi.invoker, invocation)
}

// Destroy will destroy invoker
func (fi *FilterInvoker) Destroy() {
	fi.invoker.Destroy()
}
