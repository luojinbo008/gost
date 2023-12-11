package registry

import "github.com/luojinbo008/gost/common"

// Registry is the interface that wraps Register、UnRegister、Subscribe and UnSubscribe method.
type Registry interface {
	common.Node

	Register(url *common.URL) error

	UnRegister(url *common.URL) error

	Subscribe(*common.URL, NotifyListener) error

	UnSubscribe(*common.URL, NotifyListener) error
}

// nolint
type NotifyListener interface {
	Notify(*ServiceEvent)

	NotifyAll([]*ServiceEvent, func())
}

// Listener Deprecated!
type Listener interface {
	Next() (*ServiceEvent, error)

	Close()
}
