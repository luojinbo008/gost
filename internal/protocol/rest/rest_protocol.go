package rest

import (
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
)

const (
	// GRPC module name
	REST = "rest"
)

func init() {
	extension.SetProtocol(REST, GetProtocol)
}

var restProtocol *RestProtocol

type RestProtocol struct {
	protocol.BaseProtocol
	serverMap  map[string]*GoRestfulServer
	serverLock sync.Mutex
}

func NewGRPCProtocol() *RestProtocol {
	return &RestProtocol{
		BaseProtocol: protocol.NewBaseProtocol(),
		serverMap:    make(map[string]*GoRestfulServer),
	}
}

func (rp *RestProtocol) Export(invoker protocol.Invoker) protocol.Exporter {
	url := invoker.GetURL()
	serviceKey := url.ServiceKey()

	exporter := NewRestExporter(serviceKey, invoker, rp.ExporterMap())
	rp.SetExporterMap(serviceKey, exporter)

	logger.Infof("[REST Protocol] Export serviceKey: %s,service: %s", serviceKey, url.String())

	rp.openServer(url)
	return exporter
}

func (rp *RestProtocol) openServer(url *common.URL) {
	rp.serverLock.Lock()
	defer rp.serverLock.Unlock()

	if _, ok := rp.serverMap[url.Location]; ok {
		return
	}

	if _, ok := rp.ExporterMap().Load(url.ServiceKey()); !ok {
		panic("[restProtocol]" + url.Key() + "is not existing")
	}

	srv := NewGoRestfulServer()
	rp.serverMap[url.Location] = srv
	srv.Start(url)
}

func (rp *RestProtocol) Refer(url *common.URL) protocol.Invoker {
	client, err := NewClient(url)

	if err != nil {
		logger.Warnf("can't dial the server: %s", url.Key())
		return nil
	}
	invoker := NewRestInvoker(url, client)
	rp.SetInvokers(invoker)
	logger.Infof("[GRPC Protcol] Refer service: %s", url.String())
	return invoker
}

func (rp *RestProtocol) Destroy() {
	logger.Infof("RestProtocol destroy.")

	rp.serverLock.Lock()
	defer rp.serverLock.Unlock()
	// for key, server := range gp.serverMap {
	// 	delete(gp.serverMap, key)
	// 	server.GracefulStop()
	// }

	rp.BaseProtocol.Destroy()
}

// GetProtocol gets gRPC protocol, will create if null.
func GetProtocol() protocol.Protocol {
	if restProtocol == nil {
		restProtocol = NewGRPCProtocol()
	}
	return restProtocol
}
