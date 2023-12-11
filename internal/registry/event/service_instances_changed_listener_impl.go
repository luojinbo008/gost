package event

import (
	"reflect"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/registry"
	"github.com/luojinbo008/gost/internal/remoting"
	gxset "github.com/luojinbo008/gost/utils/container/set"
	"github.com/luojinbo008/gost/utils/gof/observer"
)

// ServiceInstancesChangedListenerImpl The Service Discovery Changed  Event Listener
type ServiceInstancesChangedListenerImpl struct {
	serviceNames *gxset.HashSet
	listeners    map[string]registry.NotifyListener
	serviceUrls  map[string][]*common.URL
	allInstances map[string][]registry.ServiceInstance
}

func NewServiceInstancesChangedListener(services *gxset.HashSet) registry.ServiceInstancesChangedListener {
	return &ServiceInstancesChangedListenerImpl{
		serviceNames: services,
		listeners:    make(map[string]registry.NotifyListener),
		serviceUrls:  make(map[string][]*common.URL),
		allInstances: make(map[string][]registry.ServiceInstance),
	}
}

// OnEvent on ServiceInstancesChangedEvent the service instances change event
func (lstn *ServiceInstancesChangedListenerImpl) OnEvent(e observer.Event) error {
	ce, ok := e.(*registry.ServiceInstancesChangedEvent)

	if !ok {
		return nil
	}

	lstn.allInstances[ce.ServiceName] = ce.Instances
	newServiceURLs := make(map[string][]*common.URL)
	for _, instance := range ce.Instances {
		newServiceURLs[ce.ServiceName] = append(newServiceURLs[ce.ServiceName], instance.ToURL())
	}
	lstn.serviceUrls = newServiceURLs

	// 变更通知
	for key, notifyListener := range lstn.listeners {
		urls := lstn.serviceUrls[key]
		events := make([]*registry.ServiceEvent, 0, len(urls))
		for _, url := range urls {
			events = append(events, &registry.ServiceEvent{
				Action:  remoting.EventTypeAdd,
				Service: url,
			})
		}

		notifyListener.NotifyAll(events, func() {})
	}

	return nil
}

// AddListenerAndNotify add notify listener and notify to listen service event
func (lstn *ServiceInstancesChangedListenerImpl) AddListenerAndNotify(serviceKey string, notify registry.NotifyListener) {
	lstn.listeners[serviceKey] = notify
	urls := lstn.serviceUrls[serviceKey]
	for _, url := range urls {
		notify.Notify(&registry.ServiceEvent{
			Action:  remoting.EventTypeAdd,
			Service: url,
		})
	}
}

// RemoveListener remove notify listener
func (lstn *ServiceInstancesChangedListenerImpl) RemoveListener(serviceKey string) {
	delete(lstn.listeners, serviceKey)
}

// GetServiceNames return all listener service names
func (lstn *ServiceInstancesChangedListenerImpl) GetServiceNames() *gxset.HashSet {
	return lstn.serviceNames
}

// Accept return true if the name is the same
func (lstn *ServiceInstancesChangedListenerImpl) Accept(e observer.Event) bool {
	if ce, ok := e.(*registry.ServiceInstancesChangedEvent); ok {
		return lstn.serviceNames.Contains(ce.ServiceName)
	}
	return false
}

// GetPriority returns -1, it will be the first invoked listener
func (lstn *ServiceInstancesChangedListenerImpl) GetPriority() int {
	return -1
}

// GetEventType returns ServiceInstancesChangedEvent
func (lstn *ServiceInstancesChangedListenerImpl) GetEventType() reflect.Type {
	return reflect.TypeOf(&registry.ServiceInstancesChangedEvent{})
}
