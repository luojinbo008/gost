package router

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/protocol"
)

type Chain interface {
	Route(*common.URL, protocol.Invocation) []protocol.Invoker
	// Refresh invokers
	SetInvokers([]protocol.Invoker)
	// AddRouters Add routers
	AddRouters([]Router)
}
