package protocol

import (
	"sync"

	"github.com/luojinbo008/gost/log/logger"
)

type Exporter interface {
	GetInvoker() Invoker
	UnExport()
}

// BaseExporter is default exporter implement.
type BaseExporter struct {
	key         string
	invoker     Invoker
	exporterMap *sync.Map
}

// NewBaseExporter creates a new BaseExporter
func NewBaseExporter(key string, invoker Invoker, exporterMap *sync.Map) *BaseExporter {
	return &BaseExporter{
		key:         key,
		invoker:     invoker,
		exporterMap: exporterMap,
	}
}

// GetInvoker gets invoker
func (de *BaseExporter) GetInvoker() Invoker {
	return de.invoker
}

// UnExport un export service.
func (de *BaseExporter) UnExport() {
	logger.Infof("Exporter unexport.")
	de.invoker.Destroy()
	de.exporterMap.Delete(de.key)
}
