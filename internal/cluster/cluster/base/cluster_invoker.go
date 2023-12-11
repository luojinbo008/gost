package base

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/cluster/loadbalance"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"

	perrors "github.com/pkg/errors"
	"go.uber.org/atomic"
)

type BaseClusterInvoker struct {
	Directory      directory.Directory
	AvailableCheck bool
	Destroyed      *atomic.Bool
	// StickyInvoker  protocol.Invoker
}

func NewBaseClusterInvoker(directory directory.Directory) BaseClusterInvoker {
	return BaseClusterInvoker{
		Directory:      directory,
		AvailableCheck: true,
		Destroyed:      atomic.NewBool(false),
	}
}

func (invoker *BaseClusterInvoker) GetURL() *common.URL {
	return invoker.Directory.GetURL()
}

func (invoker *BaseClusterInvoker) Destroy() {
	// this is must atom operation
	if invoker.Destroyed.CAS(false, true) {
		invoker.Directory.Destroy()
	}
}

func (invoker *BaseClusterInvoker) IsAvailable() bool {
	return invoker.Directory.IsAvailable()
}

// CheckInvokers checks invokers' status if is available or not
func (invoker *BaseClusterInvoker) CheckInvokers(invokers []protocol.Invoker, invocation protocol.Invocation) error {
	if len(invokers) == 0 {
		ip := common.GetLocalIp()
		return perrors.Errorf("Failed to invoke the method %v. No provider available for the service %v from "+
			"registry %v on the consumer %v using the mudu version %v .Please check if the providers have been started and registered.",
			invocation.MethodName(), invoker.Directory.GetURL().SubURL.Key(), invoker.Directory.GetURL().String(), ip, constant.VersionKey)
	}
	return nil
}

// CheckWhetherDestroyed checks if cluster invoker was destroyed or not
func (invoker *BaseClusterInvoker) CheckWhetherDestroyed() error {
	if invoker.Destroyed.Load() {
		ip := common.GetLocalIp()
		return perrors.Errorf("Rpc cluster invoker for %v on consumer %v use mudu version %v is now destroyed! can not invoke any more. ",
			invoker.Directory.GetURL().Service(), ip, constant.VersionKey)
	}
	return nil
}

func (invoker *BaseClusterInvoker) DoSelect(lb loadbalance.LoadBalance, invocation protocol.Invocation, invokers []protocol.Invoker, invoked []protocol.Invoker) protocol.Invoker {
	var selectedInvoker protocol.Invoker
	if len(invokers) <= 0 {
		return selectedInvoker
	}
	selectedInvoker = invoker.doSelectInvoker(lb, invocation, invokers, invoked)
	return selectedInvoker
}

func (invoker *BaseClusterInvoker) doSelectInvoker(lb loadbalance.LoadBalance, invocation protocol.Invocation, invokers []protocol.Invoker, invoked []protocol.Invoker) protocol.Invoker {
	if len(invokers) == 0 {
		return nil
	}

	if len(invokers) == 1 {
		if invokers[0].IsAvailable() {
			return invokers[0]
		}
		protocol.SetInvokerUnhealthyStatus(invokers[0])
		logger.Errorf("the invokers of %s is nil. ", invokers[0].GetURL().ServiceKey())
		return nil
	}

	selectedInvoker := lb.Select(invokers, invocation)

	// judge if the selected Invoker is invoked and available
	if (!selectedInvoker.IsAvailable() && invoker.AvailableCheck) || isInvoked(selectedInvoker, invoked) {
		protocol.SetInvokerUnhealthyStatus(selectedInvoker)
		otherInvokers := getOtherInvokers(invokers, selectedInvoker)
		// do reselect
		for i := 0; i < 3; i++ {
			if len(otherInvokers) == 0 {
				// no other ivk to reselect, return to fallback
				break
			}
			reselectedInvoker := lb.Select(otherInvokers, invocation)
			if isInvoked(reselectedInvoker, invoked) {
				otherInvokers = getOtherInvokers(otherInvokers, reselectedInvoker)
				continue
			}
			if !reselectedInvoker.IsAvailable() {
				logger.Infof("the invoker of %s is not available, maybe some network error happened or the server is shutdown.",
					invoker.GetURL().Ip)
				protocol.SetInvokerUnhealthyStatus(reselectedInvoker)
				otherInvokers = getOtherInvokers(otherInvokers, reselectedInvoker)
				continue
			}
			return reselectedInvoker
		}
	}

	return selectedInvoker
}

func isInvoked(selectedInvoker protocol.Invoker, invoked []protocol.Invoker) bool {
	for _, i := range invoked {
		if i == selectedInvoker {
			return true
		}
	}
	return false
}

func GetLoadBalance(invoker protocol.Invoker, methodName string) loadbalance.LoadBalance {
	url := invoker.GetURL()

	// Get the service loadbalance config
	lb := url.GetParam(constant.LoadbalanceKey, constant.DefaultLoadBalance)

	return extension.GetLoadbalance(lb)
}

func getOtherInvokers(invokers []protocol.Invoker, invoker protocol.Invoker) []protocol.Invoker {
	otherInvokers := make([]protocol.Invoker, 0)
	for _, i := range invokers {
		if i != invoker {
			otherInvokers = append(otherInvokers, i)
		}
	}
	return otherInvokers
}
