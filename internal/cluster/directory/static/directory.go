package static

import (
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/cluster/directory/base"
	"github.com/luojinbo008/gost/internal/cluster/router/chain"
	"github.com/luojinbo008/gost/internal/protocol"

	perrors "github.com/pkg/errors"
)

type directory struct {
	base.Directory
	invokers []protocol.Invoker
}

// NewDirectory Create a new staticDirectory with invokers
func NewDirectory(invokers []protocol.Invoker) *directory {
	var url *common.URL

	if len(invokers) > 0 {
		url = invokers[0].GetURL()
	}
	dir := &directory{
		Directory: base.NewDirectory(url),
		invokers:  invokers,
	}

	dir.RouterChain().SetInvokers(invokers)
	return dir
}

// for-loop invokers ,if all invokers is available ,then it means directory is available
func (dir *directory) IsAvailable() bool {
	if len(dir.invokers) == 0 {
		return false
	}
	for _, invoker := range dir.invokers {
		if !invoker.IsAvailable() {
			return false
		}
	}
	return true
}

// List List invokers
func (dir *directory) List(invocation protocol.Invocation) []protocol.Invoker {
	l := len(dir.invokers)
	invokers := make([]protocol.Invoker, l)
	copy(invokers, dir.invokers)
	routerChain := dir.RouterChain()

	if routerChain == nil {
		return invokers
	}
	dirUrl := dir.GetURL()
	return routerChain.Route(dirUrl, invocation)
}

// Destroy Destroy
func (dir *directory) Destroy() {
	dir.Directory.Destroy(func() {
		for _, ivk := range dir.invokers {
			ivk.Destroy()
		}
		dir.invokers = []protocol.Invoker{}
	})
}

// BuildRouterChain build router chain by invokers
func (dir *directory) BuildRouterChain(invokers []protocol.Invoker) error {
	if len(invokers) == 0 {
		return perrors.Errorf("invokers == null")
	}
	routerChain, e := chain.NewRouterChain()
	if e != nil {
		return e
	}
	routerChain.SetInvokers(dir.invokers)
	dir.SetRouterChain(routerChain)
	return nil
}
