package config

import (
	"sync"

	"github.com/pkg/errors"
)

var (
	startOnce sync.Once
)

// RootConfig is the root config
type RootConfig struct {
	Application *ApplicationConfig `yaml:"application" json:"application"`

	Protocols map[string]*ProtocolConfig `validate:"required" yaml:"protocols" json:"protocols"`

	Provider *ProviderConfig `yaml:"provider" json:"provider"`

	Registries map[string]*RegistryConfig `yaml:"registries" json:"registries"`
}

func (rc *RootConfig) Init() error {
	// init application
	if err := rc.Application.Init(); err != nil {
		return err
	}

	// init protocol
	protocols := rc.Protocols

	if len(protocols) <= 0 {
		return errors.New("no protocol config setting")
	}

	for _, protocol := range protocols {
		if err := protocol.Init(); err != nil {
			return err
		}
	}

	// init registry
	registries := rc.Registries
	if len(registries) == 0 {
		for _, reg := range registries {
			if err := reg.Init(); err != nil {
				return err
			}
		}
	}

	// provider must last init
	if err := rc.Provider.Init(rc); err != nil {
		return err
	}

	SetRootConfig(*rc)

	// todo if we can remove this from Init in the future?
	rc.Start()

	//gracefulShutdownInit()
	return nil
}

func (rc *RootConfig) Start() {
	startOnce.Do(func() {
		// todo 优雅退出
		// 加载 Provider
		rc.Provider.Load()
	})
}

// getRegistryIds get registry ids
func (rc *RootConfig) getRegistryIds() []string {
	ids := make([]string, 0)
	for key := range rc.Registries {
		ids = append(ids, key)
	}
	return removeDuplicateElement(ids)
}

func SetRootConfig(r RootConfig) {
	rootConfig = &r
}

func GetRootConfig() *RootConfig {
	return rootConfig
}

func GetApplicationConfig() *ApplicationConfig {
	if err := check(); err == nil && rootConfig.Application != nil {
		return rootConfig.Application
	}
	return NewApplicationConfigBuilder().Build()
}

func GetProviderConfig() *ProviderConfig {
	if err := check(); err == nil && rootConfig.Provider != nil {
		return rootConfig.Provider
	}
	return NewProviderConfigBuilder().Build()
}

// newEmptyRootConfig get empty root config
func newEmptyRootConfig() *RootConfig {
	newRootConfig := &RootConfig{
		Provider:   NewProviderConfigBuilder().Build(),
		Registries: make(map[string]*RegistryConfig),
		Protocols:  make(map[string]*ProtocolConfig),
	}
	return newRootConfig
}

func NewRootConfigBuilder() *RootConfigBuilder {
	return &RootConfigBuilder{
		rootConfig: newEmptyRootConfig(),
	}
}

type RootConfigBuilder struct {
	rootConfig *RootConfig
}

func (rb *RootConfigBuilder) Build() *RootConfig {
	return rb.rootConfig
}
