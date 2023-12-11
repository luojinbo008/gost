package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/protocol"

	invocation_impl "github.com/luojinbo008/gost/internal/protocol/invocation"
)

type ClientInfo struct {
	// 接口名
	InterfaceName string

	Group   string
	Version string
}

type Client struct {
	info *ClientInfo

	cliOpts *ClientOptions

	refOpts map[string]*ReferOptions
}

func (cli *Client) call(ctx context.Context, paramsRawVals []interface{}, interfaceName, methodName, callType string, opts ...CallOption) (protocol.Result, error) {
	// get a default CallOptions
	// apply CallOption
	options := newDefaultCallOptions()
	for _, opt := range opts {
		opt(options)
	}

	inv, err := generateInvocation(methodName, paramsRawVals, callType, options)
	if err != nil {
		return nil, err
	}

	refOption := cli.refOpts[common.ServiceKey(interfaceName, options.Group, options.Version)]

	if refOption == nil {
		return nil, fmt.Errorf("no service found for %s/%s:%s, please check if the service has been registered", options.Group, interfaceName, options.Version)
	}

	return refOption.invoker.Invoke(ctx, inv), nil

}

func (cli *Client) CallUnary(ctx context.Context, req, resp interface{}, interfaceName, methodName string, opts ...CallOption) error {
	res, err := cli.call(ctx, []interface{}{req, resp}, interfaceName, methodName, "123", opts...)
	if err != nil {
		return err
	}
	return res.Error()
}

func generateInvocation(methodName string, paramsRawVals []interface{}, callType string, opts *CallOptions) (protocol.Invocation, error) {
	inv := invocation_impl.NewRPCInvocationWithOptions(
		invocation_impl.WithMethodName(methodName),
		invocation_impl.WithAttachment(constant.TimeoutKey, opts.RequestTimeout),
		invocation_impl.WithArguments(paramsRawVals),
	)
	// inv.SetAttribute(constant.CallTypeKey, callType)
	return inv, nil
}

// 初始化
func (cli *Client) Init(info *ClientInfo, opts ...ReferOption) (string, string, error) {
	if info == nil {
		return "", "", errors.New("ClientInfo is nil")
	}
	newRefOptions := defaultReferptions()
	err := newRefOptions.init(cli, opts...)
	if err != nil {
		return "", "", err
	}
	cli.refOpts[common.ServiceKey(info.InterfaceName, info.Group, info.Version)] = newRefOptions
	newRefOptions.ReferWithInfo(info)

	return info.Group, info.Version, nil
}

func NewClient(opts ...ClientOption) (*Client, error) {
	newCliOpts := defaultClientOptions()
	if err := newCliOpts.init(opts...); err != nil {
		return nil, err
	}

	return &Client{
		cliOpts: newCliOpts,
		refOpts: make(map[string]*ReferOptions),
	}, nil
}
