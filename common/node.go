package common

// Node use for process node
type Node interface {
	GetURL() *URL
	IsAvailable() bool
	Destroy()
}
