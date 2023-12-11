package __

import (
	"context"

	"github.com/luojinbo008/gost/service"
	"google.golang.org/grpc"
)

func init() {
	service.SetConsumerService(&GrpcGreeterImpl{})
}

type GrpcGreeterImpl struct {
	SayHello func(ctx context.Context, in *HelloRequest, out *HelloReply) error
}

// Reference ...
func (u *GrpcGreeterImpl) Reference() string {
	return "GrpcGreeterImpl"
}

// GetGostStub ...
func (u *GrpcGreeterImpl) GetGostStub(cc *grpc.ClientConn) GreeterClient {
	return NewGreeterClient(cc)
}
