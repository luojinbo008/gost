package nacos

import (
	"bytes"
	"net/url"
	"reflect"
	"strconv"
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/registry"
	"github.com/luojinbo008/gost/internal/remoting"
	"github.com/luojinbo008/gost/log/logger"
	gxchan "github.com/luojinbo008/gost/utils/container/chan"
	nacosClient "github.com/luojinbo008/gost/utils/database/kv/nacos"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	perrors "github.com/pkg/errors"
)

var (
	listenerCache sync.Map
)

type callback func(services []model.Instance, err error)

type nacosListener struct {
	namingClient   *nacosClient.NacosNamingClient
	listenURL      *common.URL
	regURL         *common.URL
	events         *gxchan.UnboundedChan
	instanceMap    map[string]model.Instance
	cacheLock      sync.Mutex
	done           chan struct{}
	subscribeParam *vo.SubscribeParam
}

// NewNacosListener creates a data listener for nacos
func NewNacosListener(url, regURL *common.URL, namingClient *nacosClient.NacosNamingClient) (*nacosListener, error) {
	listener := &nacosListener{
		namingClient: namingClient,
		listenURL:    url,
		regURL:       regURL,
		events:       gxchan.NewUnboundedChan(32),
		instanceMap:  map[string]model.Instance{},
		done:         make(chan struct{}),
	}
	err := listener.startListen()
	return listener, err
}

func generateUrl(instance model.Instance) *common.URL {
	if instance.Metadata == nil {
		logger.Errorf("nacos instance metadata is empty,instance:%+v", instance)
		return nil
	}

	// todo 需要优化
	path := instance.Metadata["path"]
	myInterface := instance.Metadata["interface"]

	if len(myInterface) == 0 {
		myInterface = instance.Metadata["app_id"]
	}

	if len(path) == 0 && len(myInterface) == 0 {
		logger.Errorf("nacos instance metadata does not have  both path key and interface key,instance:%+v", instance)
		return nil
	}

	if len(path) == 0 && len(myInterface) != 0 {
		path = "/" + myInterface
	}

	// todo 需要优化 暂时默认rest
	protocol := instance.Metadata["application_protocol"]
	if len(protocol) == 0 {
		protocol = "rest"
	}

	// if len(protocol) == 0 {
	// 	logger.Errorf("nacos instance metadata does not have protocol key,instance:%+v", instance)
	// 	return nil
	// }

	urlMap := url.Values{}
	for k, v := range instance.Metadata {
		urlMap.Set(k, v)
	}

	return common.NewURLWithOptions(
		common.WithIp(instance.Ip),
		common.WithPort(strconv.Itoa(int(instance.Port))),
		common.WithProtocol(protocol),
		common.WithParams(urlMap),
		common.WithPath(path),
	)
}

// Callback will be invoked when got subscribed events.
func (nl *nacosListener) Callback(services []model.Instance, err error) {
	if err != nil {
		logger.Errorf("nacos subscribe callback error:%s , subscribe:%+v ", err.Error(), nl.subscribeParam)
		return
	}

	addInstances := make([]model.Instance, 0, len(services))
	delInstances := make([]model.Instance, 0, len(services))
	updateInstances := make([]model.Instance, 0, len(services))
	newInstanceMap := make(map[string]model.Instance, len(services))

	nl.cacheLock.Lock()
	defer nl.cacheLock.Unlock()
	for i := range services {
		logger.Infof("nacos Listener Callback Instance [%v]", services[i].ClusterName)

		if !services[i].Enable {
			// instance is not available,so ignore it
			continue
		}

		host := services[i].Ip + ":" + strconv.Itoa(int(services[i].Port))
		instance := services[i]
		newInstanceMap[host] = instance

		if old, ok := nl.instanceMap[host]; !ok && instance.Healthy {
			// instance does not exist in cache, add it to cache
			addInstances = append(addInstances, instance)
		} else if !reflect.DeepEqual(old, instance) && instance.Healthy {
			// instance is not different from cache, update it to cache
			updateInstances = append(updateInstances, instance)
		}
	}

	for host, inst := range nl.instanceMap {
		if newInstance, ok := newInstanceMap[host]; !ok || !newInstance.Healthy {
			// cache instance does not exist in new instance list, remove it from cache
			delInstances = append(delInstances, inst)
		}
	}

	nl.instanceMap = newInstanceMap

	for i := range addInstances {
		if newUrl := generateUrl(addInstances[i]); newUrl != nil {
			nl.process(&remoting.ChangeEvent{Value: newUrl, EventType: remoting.EventTypeAdd})
		}
	}

	for i := range delInstances {
		if newUrl := generateUrl(delInstances[i]); newUrl != nil {
			nl.process(&remoting.ChangeEvent{Value: newUrl, EventType: remoting.EventTypeDel})
		}
	}

	for i := range updateInstances {
		if newUrl := generateUrl(updateInstances[i]); newUrl != nil {
			nl.process(&remoting.ChangeEvent{Value: newUrl, EventType: remoting.EventTypeUpdate})
		}
	}
}

func getSubscribeName(url *common.URL) string {
	var buffer bytes.Buffer
	appendParam(&buffer, url, constant.ServiceKey)
	return buffer.String()
}

func (nl *nacosListener) startListen() error {
	if nl.namingClient == nil {
		return perrors.New("nacos naming namingClient stopped")
	}
	nl.subscribeParam = createSubscribeParam(nl.listenURL, nl.regURL, nl.Callback)

	if nl.subscribeParam == nil {
		return perrors.New("create nacos subscribeParam failed")
	}

	go func() {
		err := nl.namingClient.Client().Subscribe(nl.subscribeParam)
		if err == nil {
			listenerCache.Store(nl.subscribeParam.ServiceName+nl.subscribeParam.GroupName, nl)
		}
	}()
	return nil
}

func (nl *nacosListener) stopListen() error {
	return nl.namingClient.Client().Unsubscribe(nl.subscribeParam)
}

func (nl *nacosListener) process(configType *remoting.ChangeEvent) {
	nl.events.In() <- configType
}

// Next returns the service event from nacos.
func (nl *nacosListener) Next() (*registry.ServiceEvent, error) {
	for {
		select {
		case <-nl.done:
			logger.Warnf("nacos listener is close!listenUrl:%+v", nl.listenURL)
			return nil, perrors.New("listener stopped")
		case val := <-nl.events.Out():
			e, _ := val.(*remoting.ChangeEvent)
			logger.Debugf("got nacos event %s", e)
			return &registry.ServiceEvent{Action: e.EventType, Service: e.Value.(*common.URL)}, nil
		}
	}
}

// nolint
func (nl *nacosListener) Close() {
	_ = nl.stopListen()
	close(nl.done)
}
