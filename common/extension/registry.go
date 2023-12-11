package extension

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/registry"
)

var registries = make(map[string]func(config *common.URL) (registry.Registry, error))

// SetRegistry sets the registry extension with @name
func SetRegistry(name string, v func(_ *common.URL) (registry.Registry, error)) {
	registries[name] = v
}

// GetRegistry finds the registry extension with @name
func GetRegistry(name string, config *common.URL) (registry.Registry, error) {
	if registries[name] == nil {
		panic("registry for " + name + " does not exist. ")
	}
	return registries[name](config)
}
