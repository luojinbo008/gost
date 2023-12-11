package config

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/protocol"
	protocolwrapper "github.com/luojinbo008/gost/internal/protocol/wrapper"
	"github.com/luojinbo008/gost/log/logger"

	"github.com/creasty/defaults"
	perrors "github.com/pkg/errors"
	"go.uber.org/atomic"
)

// 服务配置
type ServiceConfig struct {
	id          string
	Filter      string   `yaml:"filter" json:"filter,omitempty"`
	ProtocolIDs []string `yaml:"protocol_ids"  json:"protocol_ids,omitempty"` // multi protocolIDs support, split by ','
	Interface   string   `yaml:"interface"  json:"interface,omitempty"`
	RegistryIDs []string `yaml:"registry_ids"  json:"registry_ids,omitempty"`
	// Cluster                     string            `default:"failover" yaml:"cluster"  json:"cluster,omitempty" property:"cluster"`
	// Loadbalance                 string            `default:"random" yaml:"loadbalance"  json:"loadbalance,omitempty"  property:"loadbalance"`
	Group   string `yaml:"group"  json:"group,omitempty"`
	Version string `yaml:"version"  json:"version,omitempty"`
	// Methods                     []*MethodConfig   `yaml:"methods"  json:"methods,omitempty" property:"methods"`
	// Warmup                      string            `yaml:"warmup"  json:"warmup,omitempty"  property:"warmup"`
	// Retries                     string            `yaml:"retries"  json:"retries,omitempty" property:"retries"`
	// Serialization               string            `yaml:"serialization" json:"serialization" property:"serialization"`
	// Params                      map[string]string `yaml:"params"  json:"params,omitempty" property:"params"`
	// Token                       string            `yaml:"token" json:"token,omitempty" property:"token"`
	// AccessLog                   string            `yaml:"accesslog" json:"accesslog,omitempty" property:"accesslog"`
	// TpsLimiter                  string            `yaml:"tps.limiter" json:"tps.limiter,omitempty" property:"tps.limiter"`
	// TpsLimitInterval            string            `yaml:"tps.limit.interval" json:"tps.limit.interval,omitempty" property:"tps.limit.interval"`
	// TpsLimitRate                string            `yaml:"tps.limit.rate" json:"tps.limit.rate,omitempty" property:"tps.limit.rate"`
	// TpsLimitStrategy            string            `yaml:"tps.limit.strategy" json:"tps.limit.strategy,omitempty" property:"tps.limit.strategy"`
	// TpsLimitRejectedHandler     string            `yaml:"tps.limit.rejected.handler" json:"tps.limit.rejected.handler,omitempty" property:"tps.limit.rejected.handler"`
	// ExecuteLimit                string            `yaml:"execute.limit" json:"execute.limit,omitempty" property:"execute.limit"`
	// ExecuteLimitRejectedHandler string            `yaml:"execute.limit.rejected.handler" json:"execute.limit.rejected.handler,omitempty" property:"execute.limit.rejected.handler"`
	NotRegister bool `yaml:"not_register" json:"not_register,omitempty"`
	// ParamSign                   string            `yaml:"param.sign" json:"param.sign,omitempty" property:"param.sign"`
	Tag string `yaml:"tag" json:"tag,omitempty"`
	// TracingKey                  string            `yaml:"tracing-key" json:"tracing-key,omitempty" propertiy:"tracing-key"`

	RCProtocolsMap map[string]*ProtocolConfig

	RCRegistriesMap map[string]*RegistryConfig
	ProxyFactoryKey string

	// adaptiveService bool
	// metricsEnable   bool // whether append metrics filter to filter chain
	unexported *atomic.Bool
	exported   *atomic.Bool
	// export          bool // a flag to control whether the current service should export or not
	rpcService common.RPCService

	cacheMutex    sync.Mutex
	cacheProtocol protocol.Protocol

	//exportersLock sync.Mutex
	exporters []protocol.Exporter

	// metadataType string
}

func (s *ServiceConfig) Init(rc *RootConfig) error {

	if err := defaults.Set(s); err != nil {
		return err
	}

	s.exported = atomic.NewBool(false)

	if s.Version == "" {
		s.Version = rc.Application.Version
	}

	if s.Group == "" {
		s.Group = rc.Application.Group
	}

	s.unexported = atomic.NewBool(false)

	if len(s.RCRegistriesMap) == 0 {
		s.RCRegistriesMap = rc.Registries
	}

	if len(s.RCProtocolsMap) == 0 {
		s.RCProtocolsMap = rc.Protocols
	}

	s.RegistryIDs = translateIds(s.RegistryIDs)
	if len(s.RegistryIDs) <= 0 {
		s.RegistryIDs = rc.Provider.RegistryIDs
	}

	s.ProtocolIDs = translateIds(s.ProtocolIDs)
	if len(s.ProtocolIDs) <= 0 {
		s.ProtocolIDs = rc.Provider.ProtocolIDs
	}

	if len(s.ProtocolIDs) <= 0 {
		for k := range rc.Protocols {
			s.ProtocolIDs = append(s.ProtocolIDs, k)
		}
	}
	return verify(s)
}

// Export exports the service
func (s *ServiceConfig) Export() error {

	// TODO: delay export
	if s.unexported != nil && s.unexported.Load() {
		err := perrors.Errorf("The service %v has already unexported!", s.Interface)
		logger.Errorf(err.Error())
		return err
	}

	if s.exported != nil && s.exported.Load() {
		logger.Warnf("The service %v has already exported!", s.Interface)
		return nil
	}

	regUrls := make([]*common.URL, 0)

	if !s.NotRegister {
		regUrls = loadRegistries(s.RegistryIDs, s.RCRegistriesMap)
	}

	urlMap := s.getUrlMap()

	protocolConfigs := loadProtocol(s.ProtocolIDs, s.RCProtocolsMap)

	if len(protocolConfigs) == 0 {
		logger.Warnf("The service %v's '%v' protocols don't has right protocolConfigs, Please check your configuration center and transfer protocol ", s.Interface, s.ProtocolIDs)
		return nil
	}

	for _, proto := range protocolConfigs {
		port := proto.Port
		ivkURL := common.NewURLWithOptions(
			common.WithPath(s.Interface),
			common.WithProtocol(proto.Name),
			common.WithIp(proto.Ip),
			common.WithPort(port),
			common.WithParams(urlMap),
		)

		if len(s.Tag) > 0 {
			ivkURL.AddParam(constant.Tagkey, s.Tag)
		}

		proxyFactory := extension.GetProxyFactory(s.ProxyFactoryKey)

		if len(regUrls) > 0 {
			s.cacheMutex.Lock()
			if s.cacheProtocol == nil {
				logger.Debugf(fmt.Sprintf("First load the registry protocol, url is {%v}!", ivkURL))

				s.cacheProtocol = extension.GetProtocol(constant.RegistryProtocol)
			}

			s.cacheMutex.Unlock()

			for _, regUrl := range regUrls {
				setRegistrySubURL(ivkURL, regUrl)
				invoker := proxyFactory.GetInvoker(regUrl)

				exporter := s.cacheProtocol.Export(invoker)
				if exporter == nil {
					return perrors.New(fmt.Sprintf("Registry protocol new exporter error, registry is {%v}, url is {%v}", regUrl, ivkURL))
				}

				s.exporters = append(s.exporters, exporter)
			}
		} else {

			// 本地服务
			invoker := proxyFactory.GetInvoker(ivkURL)
			exporter := extension.GetProtocol(protocolwrapper.FILTER).Export(invoker)
			if exporter == nil {
				return perrors.New(fmt.Sprintf("Filter protocol without registry new exporter error, url is {%v}", ivkURL))
			}
			s.exporters = append(s.exporters, exporter)
		}
	}
	s.exported.Store(true)
	return nil
}

func (s *ServiceConfig) getUrlMap() url.Values {

	urlMap := url.Values{}

	urlMap.Set(constant.InterfaceKey, s.Interface)
	urlMap.Set(constant.TimestampKey, strconv.FormatInt(time.Now().Unix(), 10))

	if s.Group != "" {
		urlMap.Set(constant.GroupKey, s.Group)
	}

	if s.Version != "" {
		urlMap.Set(constant.VersionKey, s.Version)
	}

	// var filterss string
	// if s.Filter != "" {
	// 	filters = s.Filter
	// }

	// urlMap.Set(constant.ServiceFilterKey, filters)

	// filter special config
	// urlMap.Set(constant.AccessLogFilterKey, s.AccessLog)

	return urlMap
}

// loadProtocol filter protocols by ids
func loadProtocol(protocolIds []string, protocols map[string]*ProtocolConfig) []*ProtocolConfig {
	returnProtocols := make([]*ProtocolConfig, 0, len(protocols))
	for _, v := range protocolIds {
		for k, config := range protocols {
			if v == k {
				returnProtocols = append(returnProtocols, config)
			}
		}
	}
	return returnProtocols
}

// Implement only store the @s and return
func (s *ServiceConfig) Implement(rpcService common.RPCService) {
	s.rpcService = rpcService
}

func loadRegistries(registryIds []string, registries map[string]*RegistryConfig) []*common.URL {
	var registryURLs []*common.URL
	for k, registryConf := range registries {
		target := false
		if len(registryIds) == 0 || (len(registryIds) == 1 && registryIds[0] == "") {
			target = true
		} else {
			// else if user config targetRegistries
			for _, tr := range registryIds {
				if tr == k {
					target = true
					break
				}
			}
		}
		if target {
			if registryURL, err := registryConf.toURL(); err != nil {
				logger.Errorf("The registry id: %s url is invalid, error: %#v", k, err)
				panic(err)
			} else {
				registryURLs = append(registryURLs, registryURL)
			}
		}
	}

	return registryURLs
}

// setRegistrySubURL set registry sub url is ivkURl
func setRegistrySubURL(ivkURL *common.URL, regUrl *common.URL) {
	ivkURL.AddParam(constant.RegistryKey, regUrl.GetParam(constant.RegistryKey, ""))
	regUrl.SubURL = ivkURL
}
