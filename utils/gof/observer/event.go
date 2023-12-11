package observer

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Event is align with Event interface in Java.
// it's the top abstraction
// Align with 2.7.5
type Event interface {
	fmt.Stringer
	GetSource() interface{}
	GetTimestamp() time.Time
}

// BaseEvent is the base implementation of Event
// You should never use it directly
type BaseEvent struct {
	Source    interface{}
	Timestamp time.Time
}

// GetSource return the source
func (b *BaseEvent) GetSource() interface{} {
	return b.Source
}

// GetTimestamp return the Timestamp when the event is created
func (b *BaseEvent) GetTimestamp() time.Time {
	return b.Timestamp
}

// String return a human readable string representing this event
func (b *BaseEvent) String() string {
	return fmt.Sprintf("BaseEvent[source = %#v]", b.Source)
}

// NewBaseEvent create an BaseEvent instance
// and the Timestamp will be current timestamp
func NewBaseEvent(source interface{}) *BaseEvent {
	return &BaseEvent{
		Source:    source,
		Timestamp: time.Now(),
	}
}
