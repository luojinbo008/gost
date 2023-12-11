package grpc

import (
	perrors "github.com/pkg/errors"
)

type (
	// ServerConfig currently is empty struct,for future expansion
	ServerConfig struct{}

	// ClientConfig wrap client call parameters
	ClientConfig struct {
		// content type, more information refer by https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
		ContentSubType string `default:"proto" yaml:"content_sub_type" json:"content_sub_type,omitempty"`
	}
)

// GetDefaultClientConfig return grpc client default call options
func GetDefaultClientConfig() ClientConfig {
	return ClientConfig{
		ContentSubType: codecProto,
	}
}

// GetDefaultServerConfig currently return empty struct,for future expansion
func GetDefaultServerConfig() ServerConfig {
	return ServerConfig{}
}

// GetClientConfig return grpc client custom call options
func GetClientConfig() ClientConfig {
	return ClientConfig{}
}

// Validate check if custom config encoding is supported in dubbo grpc
func (c *ClientConfig) Validate() error {
	if c.ContentSubType != codecJson && c.ContentSubType != codecProto {
		return perrors.Errorf(" dubbo-go grpc codec currently only support proto、json, %s isn't supported,"+
			" please check protocol content_sub_type config", c.ContentSubType)
	}
	return nil
}

// Validate currently return empty struct,for future expansion
func (c *ServerConfig) Validate() error {
	return nil
}
