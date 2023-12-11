package config

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/log/logger"
)

// 注册配置
type RegistryConfig struct {
	Protocol  string `validate:"required" yaml:"protocol"  json:"protocol,omitempty"`
	Timeout   string `default:"5s" validate:"required" yaml:"timeout" json:"timeout,omitempty"`
	Group     string `yaml:"group" json:"group,omitempty"`
	Namespace string `yaml:"namespace" json:"namespace,omitempty"`
	TTL       string `default:"10s" yaml:"ttl" json:"ttl,omitempty"` // unit: minute
	Address   string `validate:"required" yaml:"address" json:"address,omitempty"`

	// Todo
	Zone   string            `yaml:"zone" json:"zone,omitempty" property:"zone"`
	Weight int64             `yaml:"weight" json:"weight,omitempty" property:"weight"`
	Params map[string]string `yaml:"params" json:"params,omitempty" property:"params"`
}

func (c *RegistryConfig) Init() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	return c.startRegistryConfig()
}

func (c *RegistryConfig) getUrlMap() url.Values {

	urlMap := url.Values{}
	urlMap.Set(constant.RegistryGroupKey, c.Group)
	urlMap.Set(constant.RegistryKey, c.Protocol)
	urlMap.Set(constant.RegistryTimeoutKey, c.Timeout)

	// multi registry invoker weight label for load balance
	urlMap.Set(constant.RegistryKey+"."+constant.RegistryLabelKey, strconv.FormatBool(true))
	urlMap.Set(constant.RegistryKey+"."+constant.RegistryZoneKey, c.Zone)
	urlMap.Set(constant.RegistryKey+"."+constant.WeightKey, strconv.FormatInt(c.Weight, 10))

	urlMap.Set(constant.RegistryTTLKey, c.TTL)
	urlMap.Set(constant.ClientNameKey, clientNameID(c.Protocol, c.Address))

	for k, v := range c.Params {
		urlMap.Set(k, v)
	}
	return urlMap
}

func (c *RegistryConfig) startRegistryConfig() error {
	// 解释注册地址
	c.translateRegistryAddress()
	return verify(c)
}

// translateRegistryAddress translate registry address
//
//	eg:address=nacos://127.0.0.1:8848 will return 127.0.0.1:8848 and protocol will set nacos
func (c *RegistryConfig) translateRegistryAddress() string {
	if strings.Contains(c.Address, "://") {
		u, err := url.Parse(c.Address)
		if err != nil {
			logger.Errorf("The registry url is invalid, error: %#v", err)
			panic(err)
		}
		c.Protocol = u.Scheme
		c.Address = strings.Join([]string{u.Host, u.Path}, "")
	}
	return c.Address
}

func (c *RegistryConfig) toURL() (*common.URL, error) {
	address := c.translateRegistryAddress()
	registryURLProtocol := constant.RegistryProtocol

	return common.NewURL(registryURLProtocol+"://"+address,
		common.WithParams(c.getUrlMap()),
		common.WithParamsValue(constant.RegistryKey, c.Protocol),
		common.WithParamsValue(constant.RegistryNamespaceKey, c.Namespace),
		common.WithParamsValue(constant.RegistryTimeoutKey, c.Timeout),
		common.WithLocation(c.Address),
	)
}

const (
	defaultNacosAddr       = "127.0.0.1:8848" // the default registry address of nacos
	defaultRegistryTimeout = "3s"             // the default registry timeout
)

type RegistryConfigOpt func(config *RegistryConfig) *RegistryConfig

// NewRegistryConfigWithProtocolDefaultPort New default registry config
// the input @protocol can only be:
// "nacos" with default addr "127.0.0.1:8848"
func NewRegistryConfigWithProtocolDefaultPort(protocol string) *RegistryConfig {
	switch protocol {
	case "nacos":
		return &RegistryConfig{
			Protocol: protocol,
			Address:  defaultNacosAddr,
			Timeout:  defaultRegistryTimeout,
		}
	default:
		return &RegistryConfig{
			Protocol: protocol,
		}
	}
}

// NewRegistryConfig creates New RegistryConfig with @opts
func NewRegistryConfig(opts ...RegistryConfigOpt) *RegistryConfig {
	newRegistryConfig := NewRegistryConfigWithProtocolDefaultPort("")
	for _, v := range opts {
		newRegistryConfig = v(newRegistryConfig)
	}
	return newRegistryConfig
}

// WithRegistryProtocol returns RegistryConfigOpt with given @regProtocol name
func WithRegistryProtocol(regProtocol string) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Protocol = regProtocol
		return config
	}
}

// WithRegistryAddress returns RegistryConfigOpt with given @addr registry address
func WithRegistryAddress(addr string) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Address = addr
		return config
	}
}

// WithRegistryTimeOut returns RegistryConfigOpt with given @timeout registry config
func WithRegistryTimeOut(timeout string) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Timeout = timeout
		return config
	}
}

// WithRegistryGroup returns RegistryConfigOpt with given @group registry group
func WithRegistryGroup(group string) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Group = group
		return config
	}
}

// WithRegistryWeight returns RegistryConfigOpt with given @weight registry weight flag
func WithRegistryWeight(weight int64) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Weight = weight
		return config
	}
}

// WithRegistryParams returns RegistryConfigOpt with given registry @params
func WithRegistryParams(params map[string]string) RegistryConfigOpt {
	return func(config *RegistryConfig) *RegistryConfig {
		config.Params = params
		return config
	}
}

func NewRegistryConfigBuilder() *RegistryConfigBuilder {
	return &RegistryConfigBuilder{
		registryConfig: &RegistryConfig{},
	}
}

type RegistryConfigBuilder struct {
	registryConfig *RegistryConfig
}

func (rcb *RegistryConfigBuilder) SetProtocol(protocol string) *RegistryConfigBuilder {
	rcb.registryConfig.Protocol = protocol
	return rcb
}

func (rcb *RegistryConfigBuilder) SetTimeout(timeout string) *RegistryConfigBuilder {
	rcb.registryConfig.Timeout = timeout
	return rcb
}

func (rcb *RegistryConfigBuilder) SetGroup(group string) *RegistryConfigBuilder {
	rcb.registryConfig.Group = group
	return rcb
}

func (rcb *RegistryConfigBuilder) SetAddress(address string) *RegistryConfigBuilder {
	rcb.registryConfig.Address = address
	return rcb
}

func (rcb *RegistryConfigBuilder) SetZone(zone string) *RegistryConfigBuilder {
	rcb.registryConfig.Zone = zone
	return rcb
}

func (rcb *RegistryConfigBuilder) SetWeight(weight int64) *RegistryConfigBuilder {
	rcb.registryConfig.Weight = weight
	return rcb
}

func (rcb *RegistryConfigBuilder) SetParams(params map[string]string) *RegistryConfigBuilder {
	rcb.registryConfig.Params = params
	return rcb
}

func (rcb *RegistryConfigBuilder) AddParam(key, value string) *RegistryConfigBuilder {
	if rcb.registryConfig.Params == nil {
		rcb.registryConfig.Params = make(map[string]string)
	}
	rcb.registryConfig.Params[key] = value
	return rcb
}

func (rcb *RegistryConfigBuilder) Build() *RegistryConfig {
	if err := rcb.registryConfig.Init(); err != nil {
		panic(err)
	}
	return rcb.registryConfig
}

// DynamicUpdateProperties update registry
func (c *RegistryConfig) DynamicUpdateProperties(updateRegistryConfig *RegistryConfig) {
	// if nacos's registry timeout not equal local root config's registry timeout , update.
	if updateRegistryConfig != nil && updateRegistryConfig.Timeout != c.Timeout {
		c.Timeout = updateRegistryConfig.Timeout
		logger.Infof("RegistryConfigs Timeout was dynamically updated, new value:%v", c.Timeout)
	}
}
