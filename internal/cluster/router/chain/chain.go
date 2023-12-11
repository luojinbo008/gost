package chain

import (
	"sort"
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/cluster/router"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"

	perrors "github.com/pkg/errors"
	"go.uber.org/atomic"
)

// RouterChain Router chain
type RouterChain struct {

	// Full list of addresses from registry, classified by method name.
	invokers []protocol.Invoker

	// Containing all routers, reconstruct every time 'route://' urls change.
	routers []router.Router

	// Fixed router instances: ConfigConditionRouter, TagRouter, e.g., the rule for each instance may change but the
	// instance will never delete or recreate.
	builtinRouters []router.Router

	mutex sync.RWMutex
}

// Route Loop routers in RouterChain and call Route method to determine the target invokers list.
func (c *RouterChain) Route(url *common.URL, invocation protocol.Invocation) []protocol.Invoker {
	finalInvokers := make([]protocol.Invoker, 0, len(c.invokers))
	// multiple invoker may include different methods, find correct invoker otherwise
	// will return the invoker without methods
	for _, invoker := range c.invokers {
		if invoker.GetURL().ServiceKey() == url.ServiceKey() {
			finalInvokers = append(finalInvokers, invoker)
		}
	}

	if len(finalInvokers) == 0 {
		finalInvokers = c.invokers
	}

	for _, r := range c.copyRouters() {
		finalInvokers = r.Route(finalInvokers, url, invocation)
	}
	return finalInvokers
}

// AddRouters Add routers to router chain
// New a array add builtinRouters which is not sorted in RouterChain and routers
// Sort the array
// Replace router array in RouterChain
func (c *RouterChain) AddRouters(routers []router.Router) {
	newRouters := make([]router.Router, 0, len(c.builtinRouters)+len(routers))
	newRouters = append(newRouters, c.builtinRouters...)
	newRouters = append(newRouters, routers...)
	sortRouter(newRouters)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.routers = newRouters
}

// SetInvokers receives updated invokers from registry center. If the times of notification exceeds countThreshold and
// time interval exceeds timeThreshold since last cache update, then notify to update the cache.
func (c *RouterChain) SetInvokers(invokers []protocol.Invoker) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.invokers = invokers
	for _, v := range c.routers {
		v.Notify(c.invokers)
	}
}

// copyRouters make a snapshot copy from RouterChain's router list.
func (c *RouterChain) copyRouters() []router.Router {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ret := make([]router.Router, 0, len(c.routers))
	ret = append(ret, c.routers...)
	return ret
}

// copyInvokers copies a snapshot of the received invokers.
func (c *RouterChain) copyInvokers() []protocol.Invoker {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if c.invokers == nil || len(c.invokers) == 0 {
		return nil
	}
	ret := make([]protocol.Invoker, 0, len(c.invokers))
	ret = append(ret, c.invokers...)
	return ret
}

// NewRouterChain init router chain
// Loop routerFactories and call NewRouter method
func NewRouterChain() (*RouterChain, error) {
	routerFactories := extension.GetRouterFactories()
	if len(routerFactories) == 0 {
		return nil, perrors.Errorf("No routerFactory exits , create one please")
	}

	routers := make([]router.Router, 0, len(routerFactories))

	for key, routerFactory := range routerFactories {
		r, err := routerFactory().NewRouter()
		if err != nil {
			logger.Errorf("Build router chain failed with routerFactories key:%s and error:%v", key, err)
			continue
		} else if r == nil {
			continue
		}
		routers = append(routers, r)
	}

	newRouters := make([]router.Router, len(routers))
	copy(newRouters, routers)

	sortRouter(newRouters)

	routerNeedsUpdateInit := atomic.Bool{}
	routerNeedsUpdateInit.Store(false)

	chain := &RouterChain{
		routers:        newRouters,
		builtinRouters: routers,
	}

	return chain, nil
}

// sortRouter Sort router instance by priority with stable algorithm
func sortRouter(routers []router.Router) {
	sort.Stable(byPriority(routers))
}

// byPriority Sort by priority
type byPriority []router.Router

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].Priority() < a[j].Priority() }
