package rest

import (
	"sync"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
)

// nolint
type RestExporter struct {
	*protocol.BaseExporter
}

func NewRestExporter(key string, invoker protocol.Invoker, exporterMap *sync.Map) *RestExporter {
	return &RestExporter{
		BaseExporter: protocol.NewBaseExporter(key, invoker, exporterMap),
	}
}

func (re *RestExporter) UnExport() {
	interfaceName := re.GetInvoker().GetURL().GetParam(constant.InterfaceKey, "")
	re.BaseExporter.UnExport()
	err := common.ServiceMap.UnRegister(interfaceName, REST, re.GetInvoker().GetURL().ServiceKey())
	if err != nil {
		logger.Errorf("[RestExporter.UnExport] error: %v", err)
	}
}
