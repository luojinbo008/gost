package nacos

import (
	"sync"
	"sync/atomic"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	configClientPool     nacosConfigClientPool
	configClientPoolOnce sync.Once
)

type nacosConfigClientPool struct {
	sync.Mutex
	configClient map[string]*NacosConfigClient
}

type NacosConfigClient struct {
	name        string
	clientLock  sync.Mutex // for Client
	client      config_client.IConfigClient
	config      vo.NacosClientParam //conn config
	valid       uint32
	activeCount uint32
	share       bool
}

func initNacosConfigClientPool() {
	configClientPool.configClient = make(map[string]*NacosConfigClient)
}

func (n *NacosConfigClient) newConfigClient() error {
	client, err := clients.NewConfigClient(n.config)
	if err != nil {
		return err
	}
	n.activeCount++
	atomic.StoreUint32(&n.valid, 1)
	n.client = client
	return nil
}

// NewNacosConfigClient create config client
func NewNacosConfigClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (*NacosConfigClient, error) {

	configClient := &NacosConfigClient{
		name:        name,
		activeCount: 0,
		share:       share,
		config:      vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc},
	}
	if !share {
		return configClient, configClient.newConfigClient()
	}
	configClientPoolOnce.Do(initNacosConfigClientPool)
	configClientPool.Lock()
	defer configClientPool.Unlock()
	if client, ok := configClientPool.configClient[name]; ok {
		client.activeCount++
		return client, nil
	}

	err := configClient.newConfigClient()
	if err == nil {
		configClientPool.configClient[name] = configClient
	}
	return configClient, err
}

// Client Get NacosConfigClient
func (n *NacosConfigClient) Client() config_client.IConfigClient {
	return n.client
}

// SetClient Set NacosConfigClient
func (n *NacosConfigClient) SetClient(client config_client.IConfigClient) {
	n.clientLock.Lock()
	n.client = client
	n.clientLock.Unlock()
}

// NacosClientValid Get nacos client valid status
func (n *NacosConfigClient) NacosClientValid() bool {

	return atomic.LoadUint32(&n.valid) == 1
}

// Close close client
func (n *NacosConfigClient) Close() {
	configClientPool.Lock()
	defer configClientPool.Unlock()
	if n.client == nil {
		return
	}
	n.activeCount--
	if n.share {
		if n.activeCount == 0 {
			n.client.CloseClient()
			n.client = nil
			atomic.StoreUint32(&n.valid, 0)
			delete(configClientPool.configClient, n.name)
		}
	} else {
		n.client.CloseClient()
		n.client = nil
		atomic.StoreUint32(&n.valid, 0)
		delete(configClientPool.configClient, n.name)
	}
}
