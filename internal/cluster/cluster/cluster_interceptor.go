package cluster

import (
	"context"

	"github.com/luojinbo008/gost/internal/protocol"
)

type Interceptor interface {
	// Invoke is the core function of a cluster interceptor, it determines the process of the interceptor
	Invoke(context.Context, protocol.Invoker, protocol.Invocation) protocol.Result
}
