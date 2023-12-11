package extension

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/registry"
)

type registryDirectory func(url *common.URL, registry registry.Registry) (directory.Directory, error)

var defaultRegistry registryDirectory

// SetDefaultRegistryDirectory sets the default registryDirectory
func SetDefaultRegistryDirectory(v registryDirectory) {
	defaultRegistry = v
}

// GetDefaultRegistryDirectory finds the registryDirectory with url and registry
func GetDefaultRegistryDirectory(config *common.URL, registry registry.Registry) (directory.Directory, error) {
	if defaultRegistry == nil {
		panic("registry directory is not existing, make sure you have import the package.")
	}
	return defaultRegistry(config, registry)
}
