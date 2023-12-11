package registry

import (
	"reflect"

	gxset "github.com/luojinbo008/gost/utils/container/set"
	"github.com/luojinbo008/gost/utils/gof/observer"
)

// ServiceInstancesChangedListener is the interface of the Service Discovery Changed Event Listener
type ServiceInstancesChangedListener interface {

	// OnEvent on ServiceInstancesChangedEvent the service instances change event
	OnEvent(e observer.Event) error

	// AddListenerAndNotify add notify listener and notify to listen service event
	AddListenerAndNotify(serviceKey string, notify NotifyListener)

	// RemoveListener remove notify listener
	RemoveListener(serviceKey string)

	// GetServiceNames return all listener service names
	GetServiceNames() *gxset.HashSet

	// Accept return true if the name is the same
	Accept(e observer.Event) bool

	// GetEventType returns ServiceInstancesChangedEvent
	GetEventType() reflect.Type

	// GetPriority returns -1, it will be the first invoked listener
	GetPriority() int
}
