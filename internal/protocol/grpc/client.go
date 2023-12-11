package grpc

import (
	"reflect"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/config"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/luojinbo008/gost/service"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v2"
)

var clientConf *ClientConfig
var clientConfInitOnce sync.Once

// Client is gRPC client include client connection and invoker
type Client struct {
	*grpc.ClientConn
	invoker reflect.Value
}

// NewClient creates a new gRPC client.
func NewClient(url *common.URL) (*Client, error) {
	clientConfInitOnce.Do(clientInit)

	// If global trace instance was set, it means trace function enabled.
	// If not, will return NoopTracer.
	tracer := opentracing.GlobalTracer()
	dialOpts := make([]grpc.DialOption, 0, 4)

	// set max send and recv msg size
	maxCallRecvMsgSize := constant.DefaultMaxCallRecvMsgSize
	if recvMsgSize, err := humanize.ParseBytes(url.GetParam(constant.MaxCallRecvMsgSize, "")); err == nil && recvMsgSize > 0 {
		maxCallRecvMsgSize = int(recvMsgSize)
	}
	maxCallSendMsgSize := constant.DefaultMaxCallSendMsgSize
	if sendMsgSize, err := humanize.ParseBytes(url.GetParam(constant.MaxCallSendMsgSize, "")); err == nil && sendMsgSize > 0 {
		maxCallSendMsgSize = int(sendMsgSize)
	}

	// consumer config client connectTimeout
	//connectTimeout := config.GetConsumerConfig().ConnectTimeout

	dialOpts = append(dialOpts,
		grpc.WithBlock(),
		// todo config network timeout
		grpc.WithTimeout(time.Second*3),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer, otgrpc.LogPayloads())),
		grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer, otgrpc.LogPayloads())),
		grpc.WithDefaultCallOptions(
			grpc.CallContentSubtype(clientConf.ContentSubType),
			grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
			grpc.MaxCallSendMsgSize(maxCallSendMsgSize),
		),
	)

	// todo tls conn
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(url.Location, dialOpts...)
	if err != nil {
		logger.Errorf("grpc dial error: %v", err)
		return nil, err
	}

	key := url.GetParam(constant.InterfaceKey, "")

	impl := service.GetConsumerServiceByInterfaceName(key)

	invoker := getInvoker(impl, conn)

	return &Client{
		ClientConn: conn,
		invoker:    reflect.ValueOf(invoker),
	}, nil
}

func clientInit() {
	// load rootConfig from runtime
	rootConfig := config.GetRootConfig()

	clientConfig := GetClientConfig()
	clientConf = &clientConfig

	// check client config and decide whether to use the default config
	defer func() {
		if clientConf == nil || len(clientConf.ContentSubType) == 0 {
			defaultClientConfig := GetDefaultClientConfig()
			clientConf = &defaultClientConfig
		}
		if err := clientConf.Validate(); err != nil {
			panic(err)
		}
	}()

	if rootConfig.Application == nil {
		return
	}

	protocolConf := config.GetRootConfig().Protocols

	grpcConf := protocolConf[GRPC]
	if grpcConf == nil {
		logger.Warnf("grpcConf is nil")
		return
	}
	grpcConfByte, err := yaml.Marshal(grpcConf)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(grpcConfByte, clientConf)
	if err != nil {
		panic(err)
	}
}

func getInvoker(impl interface{}, conn *grpc.ClientConn) interface{} {
	var in []reflect.Value
	in = append(in, reflect.ValueOf(conn))
	method := reflect.ValueOf(impl).MethodByName("GetGostStub")
	res := method.Call(in)
	return res[0].Interface()
}
