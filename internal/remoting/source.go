package remoting

import "fmt"

// ChangeEvent for changing listener's event
type ChangeEvent struct {
	Key       string
	Value     interface{}
	EventType EventType
}

func (c ChangeEvent) String() string {
	return fmt.Sprintf("ChangeEvent{key = %v , value = %v , eventType = %v}", c.Key, c.Value, c.EventType)
}
