package config

import (
	"github.com/creasty/defaults"
)

// protocol 配置
type ProtocolConfig struct {
	Name   string      `validate:"required" yaml:"name" json:"name,omitempty"`
	Ip     string      `yaml:"ip"  json:"ip,omitempty"`
	Port   string      `default:"50051" yaml:"port" json:"port,omitempty"`
	Params interface{} `yaml:"params" json:"params,omitempty"`

	// MaxServerSendMsgSize max size of server send message, 1mb=1000kb=1000000b 1mib=1024kb=1048576b.
	// more detail to see https://pkg.go.dev/github.com/dustin/go-humanize#pkg-constants
	MaxServerSendMsgSize string `yaml:"max-server-send-msg-size" json:"max-server-send-msg-size,omitempty"`

	// MaxServerRecvMsgSize max size of server receive message
	MaxServerRecvMsgSize string `default:"4mib" yaml:"max-server-recv-msg-size" json:"max-server-recv-msg-size,omitempty"`
}

func (p *ProtocolConfig) Init() error {
	if err := defaults.Set(p); err != nil {
		return err
	}
	return verify(p)
}
