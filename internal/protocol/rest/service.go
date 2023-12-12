package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/config"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/service"
)

// 具体业务实现接口
type RestServer interface {

	// Start rest server
	Start(*http.Server)

	// Deploy a http api
	// Deploy(routeFunc echo.HandlerFunc, middleFunc ...echo.MiddlewareFunc)

	// Destroy rest server
	// Destroy()
}

type GoRestfulServer struct {
}

func NewGoRestfulServer() *GoRestfulServer {
	return &GoRestfulServer{}
}

func (grs *GoRestfulServer) Start(url *common.URL) {
	svr := &http.Server{
		Addr:         url.Location,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		providerServices := config.GetProviderConfig().Services
		if len(providerServices) == 0 {
			panic("provider service map is null")
		}
		registerService(providerServices, svr)
	}()
}

func registerService(providerServices map[string]*config.ServiceConfig, server *http.Server) {

	for key, providerService := range providerServices {
		serviceKey := common.ServiceKey(providerService.Interface, providerService.Group, providerService.Version)
		svr := service.GetProviderService(key)

		ds, ok := svr.(RestServer)

		if !ok {
			panic("illegal service type registered")
		}

		exporter, _ := restProtocol.ExporterMap().Load(serviceKey)

		if exporter == nil {
			panic(fmt.Sprintf("no exporter found for servicekey: %v", serviceKey))
		}
		invoker := exporter.(protocol.Exporter).GetInvoker()
		if invoker == nil {
			panic(fmt.Sprintf("no invoker found for servicekey: %v", serviceKey))
		}

		ds.Start(server)
	}
}
