package cluster

import (
	"context"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/cluster/directory"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
	perrors "github.com/pkg/errors"
)

var Count int

type Rest struct {
	Tried   int
	Success bool
}

type mockCluster struct{}

// NewMockCluster returns a mock cluster instance.
//
// Mock cluster is usually used for service degradation, such as an authentication service.
// When the service provider is completely hung up, the client does not throw an exception,
// return an authorization failure through the Mock data instead.
func NewMockCluster() Cluster {
	return &mockCluster{}
}

func (cluster *mockCluster) Join(directory directory.Directory) protocol.Invoker {
	return BuildInterceptorChain(protocol.NewBaseInvoker(directory.GetURL()))
}

type MockInvoker struct {
	url       *common.URL
	available bool
	destroyed bool

	successCount int
}

func NewMockInvoker(url *common.URL, successCount int) *MockInvoker {
	return &MockInvoker{
		url:          url,
		available:    true,
		destroyed:    false,
		successCount: successCount,
	}
}

func (bi *MockInvoker) GetURL() *common.URL {
	return bi.url
}

func (bi *MockInvoker) IsAvailable() bool {
	return bi.available
}

func (bi *MockInvoker) IsDestroyed() bool {
	return bi.destroyed
}

func (bi *MockInvoker) Invoke(c context.Context, invocation protocol.Invocation) protocol.Result {
	Count++
	var (
		success bool
		err     error
	)
	if Count >= bi.successCount {
		success = true
	} else {
		err = perrors.New("error")
	}
	result := &protocol.RPCResult{Err: err, Rest: Rest{Tried: Count, Success: success}}

	return result
}

func (bi *MockInvoker) Destroy() {
	logger.Infof("Destroy invoker: %v", bi.GetURL().String())
	bi.destroyed = true
	bi.available = false
}
