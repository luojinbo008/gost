package router

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/protocol"
)

type RouterFactory interface {
	NewRouter() (Router, error)
}

type Router interface {
	// Route Determine the target invokers list.
	Route([]protocol.Invoker, *common.URL, protocol.Invocation) []protocol.Invoker

	// URL Return URL in router
	URL() *common.URL

	// Priority Return Priority in router
	// 0 to ^int(0) is better
	Priority() int64

	// Notify the router the invoker list
	Notify(invokers []protocol.Invoker)
}
