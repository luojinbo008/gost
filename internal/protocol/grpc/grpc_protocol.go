package grpc

import (
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/extension"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
)

const (
	// GRPC module name
	GRPC = "grpc"
)

func init() {
	extension.SetProtocol(GRPC, GetProtocol)
}

var grpcProtocol *GrpcProtocol

// GrpcProtocol is gRPC protocol
type GrpcProtocol struct {
	protocol.BaseProtocol
	serverMap  map[string]*Server
	serverLock sync.Mutex
}

// NewGRPCProtocol creates new gRPC protocol
func NewGRPCProtocol() *GrpcProtocol {
	return &GrpcProtocol{
		BaseProtocol: protocol.NewBaseProtocol(),
		serverMap:    make(map[string]*Server),
	}
}

// Export gRPC service for remote invocation
func (gp *GrpcProtocol) Export(invoker protocol.Invoker) protocol.Exporter {
	url := invoker.GetURL()
	serviceKey := url.ServiceKey()

	exporter := NewGrpcExporter(serviceKey, invoker, gp.ExporterMap())
	gp.SetExporterMap(serviceKey, exporter)
	logger.Infof("[GRPC Protocol] Export serviceKey: %s,service: %s", serviceKey, url.String())
	gp.openServer(url)
	return exporter
}

func (gp *GrpcProtocol) openServer(url *common.URL) {
	gp.serverLock.Lock()
	defer gp.serverLock.Unlock()

	if _, ok := gp.serverMap[url.Location]; ok {
		return
	}

	if _, ok := gp.ExporterMap().Load(url.ServiceKey()); !ok {
		panic("[GrpcProtocol]" + url.Key() + "is not existing")
	}

	srv := NewServer()
	gp.serverMap[url.Location] = srv
	srv.Start(url)
}

// Refer a remote gRPC service
func (gp *GrpcProtocol) Refer(url *common.URL) protocol.Invoker {
	client, err := NewClient(url)
	if err != nil {
		logger.Warnf("can't dial the server: %s", url.Key())
		return nil
	}
	invoker := NewGrpcInvoker(url, client)
	gp.SetInvokers(invoker)
	logger.Infof("[GRPC Protcol] Refer service: %s", url.String())
	return invoker
}

// Destroy will destroy gRPC all invoker and exporter, so it only is called once.
func (gp *GrpcProtocol) Destroy() {
	logger.Infof("GrpcProtocol destroy.")

	gp.serverLock.Lock()
	defer gp.serverLock.Unlock()
	for key, server := range gp.serverMap {
		delete(gp.serverMap, key)
		server.GracefulStop()
	}

	gp.BaseProtocol.Destroy()
}

// GetProtocol gets gRPC protocol, will create if null.
func GetProtocol() protocol.Protocol {
	if grpcProtocol == nil {
		grpcProtocol = NewGRPCProtocol()
	}
	return grpcProtocol
}
