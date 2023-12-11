package grpc

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/config"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/luojinbo008/gost/service"

	"github.com/dustin/go-humanize"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// GrpcService is gRPC service
type GrpcService interface {
	// SetProxyImpl sets proxy.
	SetProxyImpl(impl protocol.Invoker)
	// GetProxyImpl gets proxy.
	GetProxyImpl() protocol.Invoker
	// ServiceDesc gets an RPC service's specification.
	ServiceDesc() *grpc.ServiceDesc
}

// Server is a gRPC server
type Server struct {
	grpcServer *grpc.Server
	bufferSize int
}

// NewServer creates a new server
func NewServer() *Server {
	return &Server{}
}

func (s *Server) SetBufferSize(n int) {
	s.bufferSize = n
}

// Start gRPC server with @url
func (s *Server) Start(url *common.URL) {
	var (
		addr string
		err  error
	)
	addr = url.Location
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	maxServerRecvMsgSize := constant.DefaultMaxServerRecvMsgSize
	if recvMsgSize, convertErr := humanize.ParseBytes(url.GetParam(constant.MaxServerRecvMsgSize, "")); convertErr == nil && recvMsgSize != 0 {
		maxServerRecvMsgSize = int(recvMsgSize)
	}
	maxServerSendMsgSize := constant.DefaultMaxServerSendMsgSize
	if sendMsgSize, convertErr := humanize.ParseBytes(url.GetParam(constant.MaxServerSendMsgSize, "")); err == convertErr && sendMsgSize != 0 {
		maxServerSendMsgSize = int(sendMsgSize)
	}

	// If global trace instance was set, then server tracer instance
	// can be get. If not, will return NoopTracer.
	tracer := opentracing.GlobalTracer()
	var serverOpts []grpc.ServerOption
	serverOpts = append(serverOpts,
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)),
		grpc.MaxRecvMsgSize(maxServerRecvMsgSize),
		grpc.MaxSendMsgSize(maxServerSendMsgSize),
	)

	// todo 结合tls 配置
	serverOpts = append(serverOpts, grpc.Creds(insecure.NewCredentials()))

	server := grpc.NewServer(serverOpts...)
	s.grpcServer = server

	go func() {
		providerServices := config.GetProviderConfig().Services

		if len(providerServices) == 0 {
			panic("provider service map is null")
		}
		// wait all exporter ready , then set proxy impl and grpc registerService
		waitGrpcExporter(providerServices)
		registerService(providerServices, server)
		reflection.Register(server)

		if err = server.Serve(lis); err != nil {
			logger.Errorf("server serve failed with err: %v", err)
		}
	}()
}

// getSyncMapLen get sync map len
func getSyncMapLen(m *sync.Map) int {
	length := 0

	m.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

// waitGrpcExporter wait until len(providerServices) = len(ExporterMap)
func waitGrpcExporter(providerServices map[string]*config.ServiceConfig) {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	pLen := len(providerServices)
	ta := time.NewTimer(10 * time.Second)
	defer ta.Stop()

	for {
		select {
		case <-t.C:
			mLen := getSyncMapLen(grpcProtocol.ExporterMap())
			if pLen == mLen {
				return
			}
		case <-ta.C:
			panic("wait grpc exporter timeout when start grpc server")
		}
	}
}

// registerService SetProxyImpl invoker and grpc service
func registerService(providerServices map[string]*config.ServiceConfig, server *grpc.Server) {
	for key, providerService := range providerServices {
		svr := service.GetProviderService(key)
		ds, ok := svr.(GrpcService)
		if !ok {
			panic("illegal service type registered")
		}

		serviceKey := common.ServiceKey(providerService.Interface, providerService.Group, providerService.Version)
		exporter, _ := grpcProtocol.ExporterMap().Load(serviceKey)
		if exporter == nil {
			panic(fmt.Sprintf("no exporter found for servicekey: %v", serviceKey))
		}
		invoker := exporter.(protocol.Exporter).GetInvoker()
		if invoker == nil {
			panic(fmt.Sprintf("no invoker found for servicekey: %v", serviceKey))
		}

		ds.SetProxyImpl(invoker)

		server.RegisterService(ds.ServiceDesc(), svr)
	}
}

// Stop gRPC server
func (s *Server) Stop() {
	s.grpcServer.Stop()
}

// GracefulStop gRPC server
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
