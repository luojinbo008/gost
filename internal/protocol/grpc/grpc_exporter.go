package grpc

import (
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
)

// nolint
type GrpcExporter struct {
	*protocol.BaseExporter
}

// NewGrpcExporter creates a new gRPC exporter
func NewGrpcExporter(key string, invoker protocol.Invoker, exporterMap *sync.Map) *GrpcExporter {
	return &GrpcExporter{
		BaseExporter: protocol.NewBaseExporter(key, invoker, exporterMap),
	}
}

// Unexport and unregister gRPC service from registry and memory.
func (gg *GrpcExporter) UnExport() {
	interfaceName := gg.GetInvoker().GetURL().GetParam(constant.InterfaceKey, "")
	gg.BaseExporter.UnExport()
	err := common.ServiceMap.UnRegister(interfaceName, GRPC, gg.GetInvoker().GetURL().ServiceKey())
	if err != nil {
		logger.Errorf("[GrpcExporter.UnExport] error: %v", err)
	}
}
