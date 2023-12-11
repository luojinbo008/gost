// Code generated by protoc-gen-echo. DO NOT EDIT.
// versions:
// - protoc-gen-echo v1.0.0
// - protoc          v4.25.1
// source: helloworld.proto

package __

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

type HelloworldServer interface {
	// Sends a greeting
	SayHello(req *EchoHelloRequest, e echo.Context) (*EchoHelloReply, error)
}

func RegisterHelloworldRouter(e *echo.Echo, s HelloworldServer) {
	e.GET("/helloworld/:name", func(c echo.Context) error {
		var req *EchoHelloRequest = new(EchoHelloRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		reply, err := s.SayHello(req, c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
}
