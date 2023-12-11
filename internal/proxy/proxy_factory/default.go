package proxy_factory

import (
	"context"
	"reflect"
	"strings"

	"github.com/luojinbo008/gost/log/logger"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/internal/proxy"
	perrors "github.com/pkg/errors"
)

func init() {
	extension.SetProxyFactory("default", NewDefaultProxyFactory)
}

// DefaultProxyFactory is the default proxy factory
type DefaultProxyFactory struct { // delegate ProxyFactory
}

// you can rewrite DefaultProxyFactory in extension and delegate the default proxy factory like below

//func WithDelegate(delegateProxyFactory ProxyFactory) Option {
//	return func(proxy ProxyFactory) {
//		proxy.(*DefaultProxyFactory).delegate = delegateProxyFactory
//	}
//}

// NewDefaultProxyFactory returns a proxy factory instance
func NewDefaultProxyFactory(_ ...proxy.Option) proxy.ProxyFactory {
	return &DefaultProxyFactory{}
}

// GetProxy gets a proxy
func (factory *DefaultProxyFactory) GetProxy(invoker protocol.Invoker, url *common.URL) *proxy.Proxy {
	return factory.GetAsyncProxy(invoker, nil, url)
}

// GetAsyncProxy gets a async proxy
func (factory *DefaultProxyFactory) GetAsyncProxy(invoker protocol.Invoker, callBack interface{}, url *common.URL) *proxy.Proxy {
	// create proxy
	attachments := map[string]string{}

	attachments[constant.AsyncKey] = url.GetParam(constant.AsyncKey, "false")

	attachments[constant.EnvironmentKey] = url.GetParam(constant.EnvironmentKey, "dev")

	return proxy.NewProxy(invoker, callBack, attachments)
}

// GetInvoker gets a invoker
func (factory *DefaultProxyFactory) GetInvoker(url *common.URL) protocol.Invoker {
	return &ProxyInvoker{
		BaseInvoker: *protocol.NewBaseInvoker(url),
	}
}

// ProxyInvoker is a invoker struct
type ProxyInvoker struct {
	protocol.BaseInvoker
}

// Invoke is used to call service method by invocation
func (pi *ProxyInvoker) Invoke(ctx context.Context, invocation protocol.Invocation) protocol.Result {
	result := &protocol.RPCResult{}
	result.SetAttachments(invocation.Attachments())

	// get providerUrl. The origin url may be is registry URL.
	url := getProviderURL(pi.GetURL())

	methodName := invocation.MethodName()
	proto := url.Protocol
	path := strings.TrimPrefix(url.Path, "/")
	args := invocation.Arguments()

	// get service
	svc := common.ServiceMap.GetServiceByServiceKey(proto, url.ServiceKey())
	if svc == nil {
		logger.Errorf("cannot find service [%s] in %s", path, proto)
		result.SetError(perrors.Errorf("cannot find service [%s] in %s", path, proto))
		return result
	}

	// get method
	method := svc.Method()[methodName]
	if method == nil {
		logger.Errorf("cannot find method [%s] of service [%s] in %s", methodName, path, proto)
		result.SetError(perrors.Errorf("cannot find method [%s] of service [%s] in %s", methodName, path, proto))
		return result
	}

	in := []reflect.Value{svc.Rcvr()}
	if method.CtxType() != nil {
		ctx = context.WithValue(ctx, constant.AttachmentKey, invocation.Attachments())
		in = append(in, method.SuiteContext(ctx))
	}

	// prepare argv
	if (len(method.ArgsType()) == 1 || len(method.ArgsType()) == 2 && method.ReplyType() == nil) &&
		method.ArgsType()[0].String() == "[]interface {}" {
		in = append(in, reflect.ValueOf(args))
	} else {
		for i := 0; i < len(args); i++ {
			t := reflect.ValueOf(args[i])
			if !t.IsValid() {
				at := method.ArgsType()[i]
				if at.Kind() == reflect.Ptr {
					at = at.Elem()
				}
				t = reflect.New(at)
			}
			in = append(in, t)
		}
	}

	// prepare replyv
	var replyv reflect.Value
	var retErr interface{}

	returnValues, callErr := callLocalMethod(method.Method(), in)

	if callErr != nil {
		logger.Warnf("Invoke function error: %+v, service: %#v", callErr.Error(), url)
		result.SetError(callErr)
		return result
	}

	if len(returnValues) == 1 {
		retErr = returnValues[0].Interface()
	} else {
		replyv = returnValues[0]
		retErr = returnValues[1].Interface()
	}

	if retErr != nil {
		result.SetError(retErr.(error))
		return result
	}
	if replyv.IsValid() && (replyv.Kind() != reflect.Ptr || replyv.Kind() == reflect.Ptr && replyv.Elem().IsValid()) {
		result.SetResult(replyv.Interface())
	}

	return result
}

func getProviderURL(url *common.URL) *common.URL {
	if url.SubURL == nil {
		return url
	}
	return url.SubURL
}
