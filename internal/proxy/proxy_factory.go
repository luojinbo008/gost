package proxy

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/protocol"
)

type ProxyFactory interface {
	GetProxy(invoker protocol.Invoker, url *common.URL) *Proxy

	GetAsyncProxy(invoker protocol.Invoker, callBack interface{}, url *common.URL) *Proxy

	GetInvoker(url *common.URL) protocol.Invoker
}

// Option will define a function of handling ProxyFactory
type Option func(ProxyFactory)
