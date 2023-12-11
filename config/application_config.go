package config

import (
	"github.com/creasty/defaults"
	"github.com/pkg/errors"
)

// 应用级配置
type ApplicationConfig struct {
	Name string `default:"gost.io" yaml:"name" json:"name,omitempty"`

	Group string `yaml:"group" json:"group,omitempty"`

	Version string `yaml:"version" json:"version,omitempty"`

	Environment string `yaml:"environment" json:"environment,omitempty"`
}

// Init application config and set default value
func (ac *ApplicationConfig) Init() error {
	if ac == nil {
		return errors.New("application is null")
	}

	if err := ac.check(); err != nil {
		return err
	}
	return nil
}

func (ac *ApplicationConfig) check() error {
	if err := defaults.Set(ac); err != nil {
		return err
	}
	return verify(ac)
}

func NewApplicationConfigBuilder() *ApplicationConfigBuilder {
	return &ApplicationConfigBuilder{application: &ApplicationConfig{}}
}

type ApplicationConfigBuilder struct {
	application *ApplicationConfig
}

func (acb *ApplicationConfigBuilder) SetName(name string) *ApplicationConfigBuilder {
	acb.application.Name = name
	return acb
}

func (acb *ApplicationConfigBuilder) SetVersion(version string) *ApplicationConfigBuilder {
	acb.application.Version = version
	return acb
}

func (acb *ApplicationConfigBuilder) SetEnvironment(environment string) *ApplicationConfigBuilder {
	acb.application.Environment = environment
	return acb
}

func (acb *ApplicationConfigBuilder) Build() *ApplicationConfig {
	return acb.application
}
