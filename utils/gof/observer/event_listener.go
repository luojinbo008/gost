package observer

import (
	"reflect"

	gxsort "github.com/luojinbo008/gost/utils/sort"
)

// EventListener is an new interface used to align
// It contains the Prioritized means that the listener has its priority
// Usually the priority of your custom implementation should be between [100, 9000]
// the number outside the range will be though as system reserve number
// usually implementation should be singleton
type EventListener interface {
	gxsort.Prioritizer
	// OnEvent handle this event
	OnEvent(e Event) error
	// GetEventType listen which event type
	GetEventType() reflect.Type
}

// ConditionalEventListener only handle the event which it can handle
type ConditionalEventListener interface {
	EventListener
	// Accept will make the decision whether it should handle this event
	Accept(e Event) bool
}

type ChangedNotify interface {
	Notify(e Event)
}
