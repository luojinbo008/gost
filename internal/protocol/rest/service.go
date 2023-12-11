package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/config"
	"github.com/luojinbo008/gost/internal/protocol"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/luojinbo008/gost/service"
)

type RestServer interface {

	// Start rest server
	Start(*http.Server)

	// Deploy a http api
	//Deploy(routeFunc func(c echo.Context) error)

	// Destroy rest server
	// Destroy()
}

type GoRestfulServer struct {
	srv *echo.Echo
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
		// err := grs.srv.StartServer(svr)
		// if err != nil && err != http.ErrServerClosed {
		// 	logger.Errorf("[Go Restful] http.server.Serve(addr{%s}) = err{%+v}", url.Location, err)
		// }
	}()
}

// registerService SetProxyImpl invoker and rest service
func registerService(providerServices map[string]*config.ServiceConfig, server *http.Server) {
	for key, providerService := range providerServices {
		svr := service.GetProviderService(key)
		logger.GetLogger().Info("%+v", svr)
		ds, ok := svr.(RestServer)
		if !ok {
			panic("illegal service type registered")
		}

		serviceKey := common.ServiceKey(providerService.Interface, providerService.Group, providerService.Version)

		exporter, _ := restProtocol.ExporterMap().Load(serviceKey)

		if exporter == nil {
			panic(fmt.Sprintf("no exporter found for servicekey: %v", serviceKey))
		}
		invoker := exporter.(protocol.Exporter).GetInvoker()
		if invoker == nil {
			panic(fmt.Sprintf("no invoker found for servicekey: %v", serviceKey))
		}

		// ds.SetProxyImpl(invoker)
		ds.Start(server)
		//server.RegisterService(ds.ServiceDesc(), svr)
	}
}

func (grs *GoRestfulServer) Deploy(routeFunc echo.HandlerFunc, middleFunc ...echo.MiddlewareFunc) {
	grs.srv.Add("GET", "", routeFunc, middleFunc...)
}
