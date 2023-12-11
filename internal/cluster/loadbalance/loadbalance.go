package loadbalance

import (
	"github.com/luojinbo008/gost/internal/protocol"
)

type LoadBalance interface {
	Select([]protocol.Invoker, protocol.Invocation) protocol.Invoker
}
