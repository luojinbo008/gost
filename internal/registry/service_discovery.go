package registry

import (
	"fmt"

	gxset "github.com/luojinbo008/gost/utils/container/set"
	gxpage "github.com/luojinbo008/gost/utils/hash/page"
)

const DefaultPageSize = 100

// ServiceDiscovery is the interface that wraps common operations of Service Discovery
type ServiceDiscovery interface {
	fmt.Stringer

	// Destroy will destroy the service discovery.
	// If the discovery cannot be destroy, it will return an error.
	Destroy() error

	// Register will register an instance of ServiceInstance to registry
	Register(instance ServiceInstance) error

	// Update will update the data of the instance in registry
	Update(instance ServiceInstance) error

	// Unregister will unregister this instance from registry
	Unregister(instance ServiceInstance) error

	// GetDefaultPageSize will return the default page size
	GetDefaultPageSize() int

	// GetServices will return the all service names.
	GetServices() *gxset.HashSet

	// GetInstances will return all service instances with serviceName
	GetInstances(serviceName string) []ServiceInstance

	// GetInstancesByPage will return a page containing instances of ServiceInstance
	// with the serviceName the page will start at offset
	GetInstancesByPage(serviceName string, offset int, pageSize int) gxpage.Pager

	// GetHealthyInstancesByPage will return a page containing instances of ServiceInstance.
	// The param healthy indices that the instance should be healthy or not.
	// The page will start at offset
	GetHealthyInstancesByPage(serviceName string, offset int, pageSize int, healthy bool) gxpage.Pager

	// GetRequestInstances gets all instances by the specified service names
	GetRequestInstances(serviceNames []string, offset int, requestedSize int) map[string]gxpage.Pager

	// AddListener adds a new ServiceInstancesChangedListenerImpl
	AddListener(listener ServiceInstancesChangedListener) error
}
