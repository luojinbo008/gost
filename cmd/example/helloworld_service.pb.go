// Code generated by protoc-gen-echo. DO NOT EDIT.
// versions:
// - protoc-gen-echo v1.0.0
// - protoc          v4.25.1
// source: helloworld.proto

package __

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
)

func HelloworldSayHelloBusinessHandler(req *EchoHelloRequest, c echo.Context) (EchoHelloReply, error) {
	rj, err := json.Marshal(req)
	if err != nil {
		return EchoHelloReply{}, err
	}
	fmt.Printf("Got HelloRequest is: %v\n", string(rj))
	return EchoHelloReply{}, nil
}
