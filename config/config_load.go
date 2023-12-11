package config

import (
	"github.com/pkg/errors"

	"github.com/knadh/koanf"
	"github.com/luojinbo008/gost/common/constant"
)

var (
	rootConfig = NewRootConfigBuilder().Build()
)

func Load(opts ...LoaderConfOption) error {

	// conf
	conf := NewLoaderConf(opts...)

	if conf.rc == nil {
		koan := GetConfigResolver(conf)
		if err := koan.UnmarshalWithConf(constant.GOST,
			rootConfig, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
			return err
		}
	} else {
		rootConfig = conf.rc
	}

	if err := rootConfig.Init(); err != nil {
		return err
	}

	return nil
}

func check() error {
	if rootConfig == nil {
		return errors.New("execute the config.Load() method first")
	}
	return nil
}
