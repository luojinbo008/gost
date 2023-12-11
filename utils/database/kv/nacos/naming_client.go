package nacos

import (
	"sync"
	"sync/atomic"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var (
	namingClientPool nacosClientPool
	clientPoolOnce   sync.Once
)

type nacosClientPool struct {
	sync.Mutex
	namingClient map[string]*NacosNamingClient
}

type NacosNamingClient struct {
	name        string
	clientLock  sync.Mutex // for Client
	client      naming_client.INamingClient
	config      vo.NacosClientParam //conn config
	valid       uint32
	activeCount uint32
	share       bool
}

func initNacosClientPool() {
	namingClientPool.namingClient = make(map[string]*NacosNamingClient)
}

// NewNacosNamingClient create nacos client
func NewNacosNamingClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (*NacosNamingClient, error) {

	namingClient := &NacosNamingClient{
		name:        name,
		activeCount: 0,
		share:       share,
		config:      vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc},
	}
	if !share {
		return namingClient, namingClient.newNamingClient()
	}
	clientPoolOnce.Do(initNacosClientPool)
	namingClientPool.Lock()
	defer namingClientPool.Unlock()
	if client, ok := namingClientPool.namingClient[name]; ok {
		client.activeCount++
		return client, nil
	}

	err := namingClient.newNamingClient()
	if err == nil {
		namingClientPool.namingClient[name] = namingClient
	}
	return namingClient, err
}

// newNamingClient create NamingClient
func (n *NacosNamingClient) newNamingClient() error {
	client, err := clients.NewNamingClient(n.config)
	if err != nil {
		return err
	}
	n.activeCount++
	atomic.StoreUint32(&n.valid, 1)
	n.client = client
	return nil
}

// Client Get NacosNamingClient
func (n *NacosNamingClient) Client() naming_client.INamingClient {
	return n.client
}

// SetClient Set NacosNamingClient
func (n *NacosNamingClient) SetClient(client naming_client.INamingClient) {
	n.clientLock.Lock()
	n.client = client
	n.clientLock.Unlock()
}

// NacosClientValid Get nacos client valid status
func (n *NacosNamingClient) NacosClientValid() bool {

	return atomic.LoadUint32(&n.valid) == 1
}

// Close close client
func (n *NacosNamingClient) Close() {
	namingClientPool.Lock()
	defer namingClientPool.Unlock()
	if n.client == nil {
		return
	}
	n.activeCount--
	if n.share {
		if n.activeCount == 0 {
			n.client.CloseClient()
			n.client = nil
			atomic.StoreUint32(&n.valid, 0)
			delete(namingClientPool.namingClient, n.name)
		}
	} else {
		n.client.CloseClient()
		n.client = nil
		atomic.StoreUint32(&n.valid, 0)
		delete(namingClientPool.namingClient, n.name)
	}
}
