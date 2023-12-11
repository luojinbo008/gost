// Remoting 层的使用者
package protocol

import (
	"sync"

	"github.com/luojinbo008/gost/common"
)

type Protocol interface {
	Export(invoker Invoker) Exporter

	Refer(url *common.URL) Invoker

	Destroy()
}

// BaseProtocol is default protocol implement.
type BaseProtocol struct {
	exporterMap *sync.Map
	invokers    []Invoker
}

// NewBaseProtocol creates a new BaseProtocol
func NewBaseProtocol() BaseProtocol {
	return BaseProtocol{
		exporterMap: new(sync.Map),
	}
}

// SetExporterMap set @exporter with @key to local memory.
func (bp *BaseProtocol) SetExporterMap(key string, exporter Exporter) {
	bp.exporterMap.Store(key, exporter)
}

// ExporterMap gets exporter map.
func (bp *BaseProtocol) ExporterMap() *sync.Map {
	return bp.exporterMap
}

// SetInvokers sets invoker into local memory
func (bp *BaseProtocol) SetInvokers(invoker Invoker) {
	bp.invokers = append(bp.invokers, invoker)
}

// Invokers gets all invokers
func (bp *BaseProtocol) Invokers() []Invoker {
	return bp.invokers
}

// Export is default export implement.
func (bp *BaseProtocol) Export(invoker Invoker) Exporter {
	return NewBaseExporter("base", invoker, bp.exporterMap)
}

// Refer is default refer implement.
func (bp *BaseProtocol) Refer(url *common.URL) Invoker {
	return NewBaseInvoker(url)
}

// Destroy will destroy all invoker and exporter, so it only is called once.
func (bp *BaseProtocol) Destroy() {
	// destroy invokers
	for _, invoker := range bp.invokers {
		if invoker != nil {
			invoker.Destroy()
		}
	}
	bp.invokers = []Invoker{}

	// un export exporters
	bp.exporterMap.Range(func(key, exporter interface{}) bool {
		if exporter != nil {
			exporter.(Exporter).UnExport()
		} else {
			bp.exporterMap.Delete(key)
		}
		return true
	})
}
