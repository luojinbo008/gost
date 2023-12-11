package protocol

import (
	"context"
	"fmt"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/log/logger"

	perrors "github.com/pkg/errors"
	uatomic "go.uber.org/atomic"
)

var (
	ErrClientClosed     = perrors.New("remoting client has closed")
	ErrNoReply          = perrors.New("request need @response")
	ErrDestroyedInvoker = perrors.New("request Destroyed invoker")
)

type Invoker interface {
	common.Node
	// Invoke the invocation and return result.
	Invoke(context.Context, Invocation) Result
}

// BaseInvoker provides default invoker implements Invoker
type BaseInvoker struct {
	url       *common.URL
	available uatomic.Bool
	destroyed uatomic.Bool
}

// NewBaseInvoker creates a new BaseInvoker
func NewBaseInvoker(url *common.URL) *BaseInvoker {
	ivk := &BaseInvoker{
		url: url,
	}
	ivk.available.Store(true)
	ivk.destroyed.Store(false)

	return ivk
}

// GetURL gets base invoker URL
func (bi *BaseInvoker) GetURL() *common.URL {
	return bi.url
}

// IsAvailable gets available flag
func (bi *BaseInvoker) IsAvailable() bool {
	return bi.available.Load()
}

// Invoke provides default invoker implement
func (bi *BaseInvoker) Invoke(context context.Context, invocation Invocation) Result {
	return &RPCResult{}
}

// Destroy changes available and destroyed flag
func (bi *BaseInvoker) Destroy() {
	logger.Infof("Destroy invoker: %s", bi.GetURL())
	bi.destroyed.Store(true)
	bi.available.Store(false)
}

// IsDestroyed gets destroyed flag
func (bi *BaseInvoker) IsDestroyed() bool {
	return bi.destroyed.Load()
}

func (bi *BaseInvoker) String() string {
	if bi.url != nil {
		return fmt.Sprintf("invoker{protocol: %s, host: %s:%s, path: %s}",
			bi.url.Protocol, bi.url.Ip, bi.url.Port, bi.url.Path)
	}
	return fmt.Sprintf("%#v", bi)
}
