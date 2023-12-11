package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/luojinbo008/gost/common/constant/file"
	"github.com/pkg/errors"
)

type loaderConf struct {
	suffix string      // loaderConf file extension default yaml
	path   string      // loaderConf file path default ./conf/mudugo.yaml
	delim  string      // loaderConf file delim default .
	bytes  []byte      // config bytes
	rc     *RootConfig // user provide rootConfig built by config api
	name   string      // config file name
}

func NewLoaderConf(opts ...LoaderConfOption) *loaderConf {
	configFilePath := "../conf/gost.yml"

	name, suffix := resolverFilePath(configFilePath)
	conf := &loaderConf{
		suffix: suffix,
		path:   absolutePath(configFilePath),
		delim:  ".",
		name:   name,
	}
	for _, opt := range opts {
		opt.apply(conf)
	}
	if conf.rc != nil {
		return conf
	}
	if len(conf.bytes) <= 0 {
		if bytes, err := ioutil.ReadFile(conf.path); err != nil {
			panic(err)
		} else {
			conf.bytes = bytes
		}
	}
	return conf
}

type LoaderConfOption interface {
	apply(vc *loaderConf)
}

type loaderConfigFunc func(*loaderConf)

func (fn loaderConfigFunc) apply(vc *loaderConf) {
	fn(vc)
}

// WithGenre set load config file suffix
// Deprecated: replaced by WithSuffix
func WithGenre(suffix string) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		g := strings.ToLower(suffix)
		if err := checkFileSuffix(g); err != nil {
			panic(err)
		}
		conf.suffix = g
	})
}

// WithSuffix set load config file suffix
func WithSuffix(suffix file.Suffix) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		conf.suffix = string(suffix)
	})
}

// WithPath set load config path
func WithPath(path string) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		conf.path = absolutePath(path)
		if bytes, err := ioutil.ReadFile(path); err != nil {
			panic(err)
		} else {
			conf.bytes = bytes
		}
		name, suffix := resolverFilePath(path)
		conf.suffix = suffix
		conf.name = name
	})
}

func WithRootConfig(rc *RootConfig) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		conf.rc = rc
	})
}

func WithDelim(delim string) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		conf.delim = delim
	})
}

// WithBytes set load config  bytes
func WithBytes(bytes []byte) LoaderConfOption {
	return loaderConfigFunc(func(conf *loaderConf) {
		conf.bytes = bytes
	})
}

// absolutePath get absolut path
func absolutePath(inPath string) string {
	if inPath == "$HOME" || strings.HasPrefix(inPath, "$HOME"+string(os.PathSeparator)) {
		inPath = userHomeDir() + inPath[5:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	return ""
}

// userHomeDir get gopath
func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// checkFileSuffix check file suffix
func checkFileSuffix(suffix string) error {
	for _, g := range []string{"json", "toml", "yaml", "yml"} {
		if g == suffix {
			return nil
		}
	}
	return errors.Errorf("no support file suffix: %s", suffix)
}

// resolverFilePath resolver file path
func resolverFilePath(path string) (name, suffix string) {
	paths := strings.Split(path, "/")
	fileName := strings.Split(paths[len(paths)-1], ".")
	if len(fileName) < 2 {
		return fileName[0], string(file.YAML)
	}
	return fileName[0], fileName[1]
}
