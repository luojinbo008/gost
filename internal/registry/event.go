package registry

import (
	"fmt"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/remoting"
	"github.com/luojinbo008/gost/utils/gof/observer"
)

type KeyFunc func(*common.URL) string

// ServiceEvent includes create, update, delete event
type ServiceEvent struct {
	Action  remoting.EventType
	Service *common.URL
	key     string // store the key for Service.Key()
	updated bool   // If the url is updated, such as Merged.
	KeyFunc KeyFunc
}

// String return the description of event
func (e *ServiceEvent) String() string {
	return fmt.Sprintf("ServiceEvent{Action{%s}, Path{%s}, Key{%s}}", e.Action, e.Service, e.key)
}

// Update updates the url with the merged URL. Work with Updated() can reduce the process
// of some merging URL.
func (e *ServiceEvent) Update(url *common.URL) {
	e.Service = url
	e.updated = true
}

// Updated checks if the url is updated. If the serviceEvent is updated, then it don't need
// merge url again.
func (e *ServiceEvent) Updated() bool {
	return e.updated
}

// Key generates the key for service.Key(). It is cached once.
func (e *ServiceEvent) Key() string {
	if len(e.key) > 0 {
		return e.key
	}
	if e.KeyFunc == nil {
		e.key = e.Service.GetCacheInvokerMapKey()
	} else {
		e.key = e.KeyFunc(e.Service)
	}
	return e.key
}

// ServiceInstancesChangedEvent represents service instances make some changing
type ServiceInstancesChangedEvent struct {
	observer.BaseEvent
	ServiceName string
	Instances   []ServiceInstance
}

// String return the description of the event
func (s *ServiceInstancesChangedEvent) String() string {
	return fmt.Sprintf("ServiceInstancesChangedEvent[source=%s]", s.ServiceName)
}

// NewServiceInstancesChangedEvent will create the ServiceInstanceChangedEvent instance
func NewServiceInstancesChangedEvent(serviceName string, instances []ServiceInstance) *ServiceInstancesChangedEvent {
	return &ServiceInstancesChangedEvent{
		BaseEvent: observer.BaseEvent{
			Source:    serviceName,
			Timestamp: time.Now(),
		},
		ServiceName: serviceName,
		Instances:   instances,
	}
}
