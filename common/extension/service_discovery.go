package extension

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/registry"

	perrors "github.com/pkg/errors"
)

var discoveryCreatorMap = make(map[string]func(url *common.URL) (registry.ServiceDiscovery, error), 4)

// SetServiceDiscovery will store the @creator and @name
// protocol indicate the implementation, like nacos
// the name like nacos-1...
func SetServiceDiscovery(protocol string, creator func(url *common.URL) (registry.ServiceDiscovery, error)) {
	discoveryCreatorMap[protocol] = creator
}

// GetServiceDiscovery will return the registry.ServiceDiscovery
// protocol indicate the implementation, like nacos
// the name like nacos-1...
// if not found, or initialize instance failed, it will return error.
func GetServiceDiscovery(url *common.URL) (registry.ServiceDiscovery, error) {
	protocol := url.GetParam(constant.RegistryKey, "")
	creator, ok := discoveryCreatorMap[protocol]
	if !ok {
		return nil, perrors.New("Could not find the service discovery with discovery protocol: " + protocol)
	}
	return creator(url)
}
