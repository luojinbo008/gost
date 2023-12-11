package extension

import (
	"github.com/luojinbo008/gost/internal/cluster/loadbalance"
)

var loadbalances = make(map[string]func() loadbalance.LoadBalance)

// SetLoadbalance sets the loadbalance extension with @name
// For example: round_robin
func SetLoadbalance(name string, fcn func() loadbalance.LoadBalance) {
	loadbalances[name] = fcn
}

// GetLoadbalance finds the loadbalance extension with @name
func GetLoadbalance(name string) loadbalance.LoadBalance {
	if loadbalances[name] == nil {
		panic("loadbalance for " + name + " is not existing, make sure you have import the package.")
	}

	return loadbalances[name]()
}
