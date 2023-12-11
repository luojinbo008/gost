package grpc

import (
	"context"
	"reflect"
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc/connectivity"
)

var errNoReply = errors.New("request need @response")

// nolint
type GrpcInvoker struct {
	protocol.BaseInvoker
	quitOnce    sync.Once
	clientGuard *sync.RWMutex
	client      *Client
}

func NewGrpcInvoker(url *common.URL, client *Client) *GrpcInvoker {
	return &GrpcInvoker{
		BaseInvoker: *protocol.NewBaseInvoker(url),
		clientGuard: &sync.RWMutex{},
		client:      client,
	}
}

func (gi *GrpcInvoker) setClient(client *Client) {
	gi.clientGuard.Lock()
	defer gi.clientGuard.Unlock()

	gi.client = client
}

func (gi *GrpcInvoker) getClient() *Client {
	gi.clientGuard.RLock()
	defer gi.clientGuard.RUnlock()

	return gi.client
}

// Invoke is used to call service method by invocation
func (gi *GrpcInvoker) Invoke(ctx context.Context, invocation protocol.Invocation) protocol.Result {
	var result protocol.RPCResult

	if !gi.BaseInvoker.IsAvailable() {
		// Generally, the case will not happen, because the invoker has been removed
		// from the invoker list before destroy,so no new request will enter the destroyed invoker
		logger.Warnf("this grpcInvoker is destroyed")
		result.Err = protocol.ErrDestroyedInvoker
		return &result
	}

	gi.clientGuard.RLock()
	defer gi.clientGuard.RUnlock()

	if gi.client == nil {
		result.Err = protocol.ErrClientClosed
		return &result
	}

	if !gi.BaseInvoker.IsAvailable() {
		// Generally, the case will not happen, because the invoker has been removed
		// from the invoker list before destroy,so no new request will enter the destroyed invoker
		logger.Warnf("this grpcInvoker is destroying")
		result.Err = protocol.ErrDestroyedInvoker
		return &result
	}

	if invocation.Reply() == nil {
		result.Err = errNoReply
	}

	var in []reflect.Value
	in = append(in, reflect.ValueOf(ctx))
	in = append(in, invocation.ParameterValues()...)

	methodName := invocation.MethodName()
	method := gi.client.invoker.MethodByName(methodName)
	res := method.Call(in)
	result.Rest = res[0]
	// check err
	if !res[1].IsNil() {
		result.Err = res[1].Interface().(error)
	}

	return &result
}

// IsAvailable get available status
func (gi *GrpcInvoker) IsAvailable() bool {
	client := gi.getClient()
	if client != nil {
		return gi.BaseInvoker.IsAvailable() && client.GetState() != connectivity.Shutdown
	}

	return false
}

// IsDestroyed get destroyed status
func (gi *GrpcInvoker) IsDestroyed() bool {
	client := gi.getClient()
	if client != nil {
		return gi.BaseInvoker.IsDestroyed() && client.GetState() == connectivity.Shutdown
	}

	return false
}

// Destroy will destroy gRPC's invoker and client, so it is only called once
func (gi *GrpcInvoker) Destroy() {
	gi.quitOnce.Do(func() {
		gi.BaseInvoker.Destroy()
		client := gi.getClient()
		if client != nil {
			gi.setClient(nil)
			client.Close()
		}
	})
}
