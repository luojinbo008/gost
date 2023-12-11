package config

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/pkg/errors"

	"github.com/luojinbo008/gost/common/constant/file"
	log "github.com/luojinbo008/gost/log/logger"
)

// GetConfigResolver get config resolver
func GetConfigResolver(conf *loaderConf) *koanf.Koanf {
	var (
		k   *koanf.Koanf
		err error
	)
	if len(conf.suffix) <= 0 {
		conf.suffix = string(file.YAML)
	}
	if len(conf.delim) <= 0 {
		conf.delim = "."
	}
	bytes := conf.bytes
	if len(bytes) <= 0 {
		panic(errors.New("bytes is nil,please set bytes or file path"))
	}
	k = koanf.New(conf.delim)

	switch conf.suffix {
	case "yaml", "yml":
		err = k.Load(rawbytes.Provider(bytes), yaml.Parser())
	case "json":
		err = k.Load(rawbytes.Provider(bytes), json.Parser())
	case "toml":
		err = k.Load(rawbytes.Provider(bytes), toml.Parser())
	default:
		err = errors.Errorf("no support %s file suffix", conf.suffix)
	}

	if err != nil {
		panic(err)
	}
	return resolvePlaceholder(k)
}

// resolvePlaceholder replace ${xx} with real value
func resolvePlaceholder(resolver *koanf.Koanf) *koanf.Koanf {
	m := make(map[string]interface{})
	for k, v := range resolver.All() {
		s, ok := v.(string)
		if !ok {
			continue
		}
		newKey, defaultValue := checkPlaceholder(s)
		if newKey == "" {
			continue
		}
		m[k] = resolver.Get(newKey)
		if m[k] == nil {
			m[k] = defaultValue
		}
	}
	err := resolver.Load(confmap.Provider(m, resolver.Delim()), nil)
	if err != nil {
		log.Errorf("resolvePlaceholder error %s", err)
	}
	return resolver
}

func checkPlaceholder(s string) (newKey, defaultValue string) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, file.PlaceholderPrefix) || !strings.HasSuffix(s, file.PlaceholderSuffix) {
		return
	}
	s = s[len(file.PlaceholderPrefix) : len(s)-len(file.PlaceholderSuffix)]
	indexColon := strings.Index(s, ":")
	if indexColon == -1 {
		newKey = strings.TrimSpace(s)
		return
	}
	newKey = strings.TrimSpace(s[0:indexColon])
	defaultValue = strings.TrimSpace(s[indexColon+1:])

	return
}
