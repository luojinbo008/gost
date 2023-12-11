package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/luojinbo008/gost/service"
)

// 服务提供方
type ProviderConfig struct {

	// Deprecated Register whether registration is required
	Register bool `yaml:"register" json:"register"`

	// RegistryIDs is registry ids list
	RegistryIDs []string `yaml:"registry-ids" json:"registry-ids"`

	// protocol
	ProtocolIDs []string `yaml:"protocol-ids" json:"protocol-ids"`

	// Services services
	Services map[string]*ServiceConfig `yaml:"services" json:"services,omitempty"`

	ProxyFactory string `default:"default" yaml:"proxy" json:"proxy,omitempty"`

	ConfigType map[string]string `yaml:"config_type" json:"config_type,omitempty"`
}

func (c *ProviderConfig) check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	return verify(c)
}

func (c *ProviderConfig) Init(rc *RootConfig) error {
	if c == nil {
		return nil
	}

	c.RegistryIDs = translateIds(c.RegistryIDs)

	// 如果没有配置 registerId, 取rootConfig 全部的注册中心
	if len(c.RegistryIDs) <= 0 {
		c.RegistryIDs = rc.getRegistryIds()
	}

	c.ProtocolIDs = translateIds(c.ProtocolIDs)

	for key, serviceConfig := range c.Services {
		if serviceConfig.Interface == "" {
			logger.Errorf("Service with reference = %s is not support read interface name from it."+
				"If you are not using pb serialization, please set 'interface' field in service config.", key)
			continue
		}

		if err := serviceConfig.Init(rc); err != nil {
			return err
		}
	}

	if err := c.check(); err != nil {
		return err
	}

	return nil
}

// TODO 放在 service 启动项中
func (c *ProviderConfig) Load() {

	for registeredTypeName, svr := range service.GetProviderServiceMap() {
		serviceConfig, ok := c.Services[registeredTypeName]

		if !ok {
			// service doesn't config in config file, create one with default
			logger.Fatalf("can not find service with registeredTypeName %s in configuration. Use the default configuration instead.", registeredTypeName)
			// todo 后续自动 init interface name 自动set config
		}

		serviceConfig.id = registeredTypeName
		serviceConfig.Implement(svr)

		if err := serviceConfig.Export(); err != nil {
			logger.Errorf(fmt.Sprintf("service with registeredTypeName = %s export failed! err: %#v", registeredTypeName, err))
		}
	}
}

// newEmptyProviderConfig returns ProviderConfig with default ApplicationConfig
func newEmptyProviderConfig() *ProviderConfig {
	newProviderConfig := &ProviderConfig{
		Services:    make(map[string]*ServiceConfig),
		RegistryIDs: make([]string, 8),
		ProtocolIDs: make([]string, 8),
	}
	return newProviderConfig
}

type ProviderConfigBuilder struct {
	providerConfig *ProviderConfig
}

func NewProviderConfigBuilder() *ProviderConfigBuilder {
	return &ProviderConfigBuilder{providerConfig: newEmptyProviderConfig()}
}

func (pcb *ProviderConfigBuilder) Build() *ProviderConfig {
	return pcb.providerConfig
}
