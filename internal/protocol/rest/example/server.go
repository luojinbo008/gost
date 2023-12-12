package __

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/luojinbo008/gost/internal/protocol"
)

// server is used to implement helloworld.GreeterServer.
type Dserver struct {
	e *echo.Echo
	*ProviderBase
}

func (s *Dserver) SayHello(req *EchoHelloRequest, e echo.Context) (*EchoHelloReply, error) {
	rj, err := json.Marshal(req)
	if err != nil {
		return &EchoHelloReply{}, err
	}

	fmt.Printf("Got HelloRequesoRequest is: %v\n", string(rj))
	return &EchoHelloReply{
		Code:    1,
		Message: string(rj),
	}, nil
}

func NewService() *Dserver {
	return &Dserver{
		e:            echo.New(),
		ProviderBase: &ProviderBase{},
	}
}

func (s *Dserver) Start(svr *http.Server) {
	go func() {
		if err := s.e.StartServer(svr); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	s.e.Use(middleware.Recover())

	RegisterHelloworldRouter(s.e, s)

	s.e.Any("/health", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "ok")
	})
}

type ProviderBase struct {
	proxyImpl protocol.Invoker
}

func (s *ProviderBase) Reference() string {
	return "RestGreeterImpl"
}

func (s *ProviderBase) SetProxyImpl(impl protocol.Invoker) {
	s.proxyImpl = impl
}

func (s *ProviderBase) GetProxyImpl() protocol.Invoker {
	return s.proxyImpl
}
