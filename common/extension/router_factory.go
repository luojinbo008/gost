package extension

import (
	"github.com/luojinbo008/gost/internal/cluster/router"
)

var (
	routers = make(map[string]func() router.RouterFactory)
)

// SetRouterFactory sets create router factory function with @name
func SetRouterFactory(name string, fun func() router.RouterFactory) {
	routers[name] = fun
}

// GetRouterFactory gets create router factory function by @name
func GetRouterFactory(name string) router.RouterFactory {
	if routers[name] == nil {
		panic("router_factory for " + name + " is not existing, make sure you have import the package.")
	}
	return routers[name]()
}

// GetRouterFactories gets all create router factory function
func GetRouterFactories() map[string]func() router.RouterFactory {
	return routers
}
