// 快速失败
package failfast

import (
	"context"

	"github.com/luojinbo008/gost/internal/cluster/cluster/base"
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/protocol"
)

type failfastClusterInvoker struct {
	base.BaseClusterInvoker
}

func newFailfastClusterInvoker(directory directory.Directory) protocol.Invoker {
	return &failfastClusterInvoker{
		BaseClusterInvoker: base.NewBaseClusterInvoker(directory),
	}
}

// nolint
func (invoker *failfastClusterInvoker) Invoke(ctx context.Context, invocation protocol.Invocation) protocol.Result {
	invokers := invoker.Directory.List(invocation)
	err := invoker.CheckInvokers(invokers, invocation)
	if err != nil {
		return &protocol.RPCResult{Err: err}
	}

	loadbalance := base.GetLoadBalance(invokers[0], invocation.ActualMethodName())

	err = invoker.CheckWhetherDestroyed()

	if err != nil {
		return &protocol.RPCResult{Err: err}
	}

	ivk := invoker.DoSelect(loadbalance, invocation, invokers, nil)
	return ivk.Invoke(ctx, invocation)
}
