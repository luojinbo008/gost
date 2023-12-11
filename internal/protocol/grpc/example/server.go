package __

import (
	"context"
	"net"

	"github.com/luojinbo008/gost/internal/protocol"
	log "github.com/luojinbo008/gost/log/logger"
	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type Dserver struct {
	*ProviderBase
}

func NewService() *Dserver {
	return &Dserver{
		ProviderBase: &ProviderBase{},
	}
}

// SayHello implements helloworld.GreeterServer
func (s *Dserver) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	log.Infof("Received: %v", in.GetName())
	return &HelloReply{Message: "Hello GOST " + in.GetName()}, nil
}

func (s *Dserver) Reference() string {
	return "GrpcGreeterImpl"
}

type Server struct {
	listener net.Listener
	server   *grpc.Server
}

func NewServer(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer()
	service := NewService()

	RegisterGreeterServer(server, service)

	s := Server{
		listener: listener,
		server:   server,
	}
	return &s, nil
}

func (s *Server) Start() {
	if err := s.server.Serve(s.listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

type ProviderBase struct {
	proxyImpl protocol.Invoker
}

func (s *ProviderBase) SetProxyImpl(impl protocol.Invoker) {
	s.proxyImpl = impl
}

func (s *ProviderBase) GetProxyImpl() protocol.Invoker {
	return s.proxyImpl
}

func (c *ProviderBase) ServiceDesc() *grpc.ServiceDesc {
	return &Greeter_ServiceDesc
}
