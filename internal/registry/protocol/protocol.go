package protocol

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/protocol"
	protocolwrapper "github.com/luojinbo008/gost/internal/protocol/wrapper"
	"github.com/luojinbo008/gost/internal/registry"
	"github.com/luojinbo008/gost/log/logger"
	gxset "github.com/luojinbo008/gost/utils/container/set"
)

var (
	regProtocol *registryProtocol
	once        sync.Once

	// TODO 保留字段

	reserveParams = []string{
		//constant.AppId,
		constant.MetaGroupKey,
		constant.VersionKey,
		constant.InterfaceKey,
	}
)

type registryProtocol struct {
	// Registry Map<RegistryAddress, Registry>
	registries *sync.Map

	// To solve the problem of RMI repeated exposure port conflicts,
	// the services that have been exposed are no longer exposed.
	// providerurl <--> exporter
	bounds *sync.Map
	once   sync.Once
}

func init() {
	extension.SetProtocol(constant.RegistryProtocol, GetProtocol)
}

func newRegistryProtocol() *registryProtocol {
	return &registryProtocol{
		registries: &sync.Map{},
		bounds:     &sync.Map{},
	}
}

// 获得注册器
func (proto *registryProtocol) getRegistry(registryUrl *common.URL) registry.Registry {
	var err error
	reg, loaded := proto.registries.Load(registryUrl.Location)
	if !loaded {
		reg, err = extension.GetRegistry(registryUrl.Protocol, registryUrl)
		if err != nil {
			logger.Errorf("Registry can not connect success, program is going to panic.Error message is %s", err.Error())
			panic(err)
		}
		proto.registries.Store(registryUrl.Location, reg)
	}
	return reg.(registry.Registry)
}

func getCacheKey(invoker protocol.Invoker) string {
	url := getProviderUrl(invoker)
	delKeys := gxset.NewSet("dynamic", "enabled")
	return url.CloneExceptParams(delKeys).String()
}

func getUrlToRegistry(providerUrl *common.URL, registryUrl *common.URL) *common.URL {
	if registryUrl.GetParamBool("simplified", true) {
		return providerUrl.CloneWithParams(reserveParams)
	} else {
		return filterHideKey(providerUrl)
	}
}

// filterHideKey filter the parameters that do not need to be output in url(Starting with .)
func filterHideKey(url *common.URL) *common.URL {
	// be careful params maps in url is map type
	removeSet := gxset.NewSet()
	for k := range url.GetParams() {
		if strings.HasPrefix(k, ".") {
			removeSet.Add(k)
		}
	}
	return url.CloneExceptParams(removeSet)
}

// nolint
func (proto *registryProtocol) GetRegistries() []registry.Registry {
	var rs []registry.Registry
	proto.registries.Range(func(_, v interface{}) bool {
		if r, ok := v.(registry.Registry); ok {
			rs = append(rs, r)
		}
		return true
	})
	return rs
}

// Refer provider service from registry center
func (proto *registryProtocol) Refer(url *common.URL) protocol.Invoker {
	return nil
	// registryUrl := url
	// serviceUrl := registryUrl.SubURL

	// reg := proto.getRegistry(url)
	// directory, err := extension.GetDefaultRegistryDirectory(registryUrl, reg)

	// if err != nil {
	// 	logger.Errorf("consumer service %v create registry directory error, error message is %s, and will return nil invoker!",
	// 		serviceUrl.String(), err.Error())
	// 	return nil
	// }

	// // new cluster invoker
	// clusterKey := serviceUrl.GetParam(constant.ClusterKey, constant.DefaultCluster)
	// cluster, err := extension.GetCluster(clusterKey)

	// if err != nil {
	// 	panic(err)
	// }
	// //
	// invoker := cluster.Join(directory)

	// return invoker
}

// Export provider service to registry center
func (proto *registryProtocol) Export(originInvoker protocol.Invoker) protocol.Exporter {

	// 注册配置
	registryUrl := getRegistryUrl(originInvoker)

	// 服务提供url配置
	providerUrl := getProviderUrl(originInvoker)

	// export invoker
	exporter := proto.doLocalExport(originInvoker, providerUrl)

	if len(registryUrl.Protocol) > 0 {

		// url to registry
		reg := proto.getRegistry(registryUrl)
		registeredProviderUrl := getUrlToRegistry(providerUrl, registryUrl)

		err := reg.Register(registeredProviderUrl)
		if err != nil {
			logger.Errorf("provider service %v register registry %v error, error message is %s",
				providerUrl.Key(), registryUrl.Key(), err.Error())
			return nil
		}
		exporter.SetRegisterUrl(registeredProviderUrl)
	} else {
		logger.Warnf("provider service %v do not regist to registry %v. possible direct connection provider",
			providerUrl.Key(), registryUrl.Key())
	}

	return exporter
}

func (proto *registryProtocol) doLocalExport(originInvoker protocol.Invoker, providerUrl *common.URL) *exporterChangeableWrapper {
	key := getCacheKey(originInvoker)
	cachedExporter, loaded := proto.bounds.Load(key)
	if !loaded {
		// new Exporter
		invokerDelegate := newInvokerDelegate(originInvoker, providerUrl)
		cachedExporter = newExporterChangeableWrapper(originInvoker,
			extension.GetProtocol(protocolwrapper.FILTER).Export(invokerDelegate))

		proto.bounds.Store(key, cachedExporter)
	}
	return cachedExporter.(*exporterChangeableWrapper)
}

// Destroy registry protocol
func (proto *registryProtocol) Destroy() {
	proto.bounds.Range(func(key, value interface{}) bool {
		// protocol holds the exporters actually, instead, registry holds them in order to avoid export repeatedly, so
		// the work for unexport should be finished in protocol.UnExport(), see also config.destroyProviderProtocols().
		exporter := value.(*exporterChangeableWrapper)
		reg := proto.getRegistry(getRegistryUrl(exporter.originInvoker))
		if err := reg.UnRegister(exporter.registerUrl); err != nil {
			panic(err)
		}
		// TODO unsubscribeUrl

		// TODO 异步推出 目前强制 15s 后退出
		go func() {
			select {
			case <-time.After(15 * time.Second):
				exporter.UnExport()
				proto.bounds.Delete(key)
			}
		}()
		return true
	})

	proto.registries.Range(func(key, value interface{}) bool {
		proto.registries.Delete(key)
		return true
	})
}

func getRegistryUrl(invoker protocol.Invoker) *common.URL {
	// here add * for return a new url
	url := invoker.GetURL()
	// if the protocol == registry, set protocol the registry value in url.params
	if url.Protocol == constant.RegistryProtocol {
		url.Protocol = url.GetParam(constant.RegistryKey, "")
	}
	return url
}

func getProviderUrl(invoker protocol.Invoker) *common.URL {
	url := invoker.GetURL()
	// be careful params maps in url is map type
	return url.SubURL.Clone()
}

func setProviderUrl(regURL *common.URL, providerURL *common.URL) {
	regURL.SubURL = providerURL
}

// GetProtocol return the singleton registryProtocol
func GetProtocol() protocol.Protocol {
	once.Do(func() {
		regProtocol = newRegistryProtocol()
	})
	return regProtocol
}

type invokerDelegate struct {
	invoker protocol.Invoker
	protocol.BaseInvoker
}

func newInvokerDelegate(invoker protocol.Invoker, url *common.URL) *invokerDelegate {
	return &invokerDelegate{
		invoker:     invoker,
		BaseInvoker: *protocol.NewBaseInvoker(url),
	}
}

// Invoke remote service base on URL of wrappedInvoker
func (ivk *invokerDelegate) Invoke(ctx context.Context, invocation protocol.Invocation) protocol.Result {
	return ivk.invoker.Invoke(ctx, invocation)
}

type exporterChangeableWrapper struct {
	protocol.Exporter
	originInvoker protocol.Invoker
	exporter      protocol.Exporter
	registerUrl   *common.URL
	subscribeUrl  *common.URL
}

func (e *exporterChangeableWrapper) UnExport() {
	e.exporter.UnExport()
}

func (e *exporterChangeableWrapper) SetRegisterUrl(registerUrl *common.URL) {
	e.registerUrl = registerUrl
}

func (e *exporterChangeableWrapper) SetSubscribeUrl(subscribeUrl *common.URL) {
	e.subscribeUrl = subscribeUrl
}

func (e *exporterChangeableWrapper) GetInvoker() protocol.Invoker {
	return e.exporter.GetInvoker()
}

func newExporterChangeableWrapper(originInvoker protocol.Invoker, exporter protocol.Exporter) *exporterChangeableWrapper {
	return &exporterChangeableWrapper{
		originInvoker: originInvoker,
		exporter:      exporter,
	}
}
