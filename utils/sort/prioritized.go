package gxsort

// Prioritizer is the abstraction of priority.
type Prioritizer interface {
	// GetPriority will return the priority
	GetPriority() int
}