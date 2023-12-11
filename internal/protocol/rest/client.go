package rest

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/luojinbo008/gost/common"
	"github.com/luojinbo008/gost/common/constant"
	"github.com/luojinbo008/gost/log/logger"
	"github.com/luojinbo008/gost/service"
)

// Client is gRPC client include client connection and invoker
type Client struct {
	*http.Client
	invoker reflect.Value
}

// NewClient creates a new gRPC client.
func NewClient(url *common.URL) (*Client, error) {

	// If global trace instance was set, it means trace function enabled.
	// If not, will return NoopTracer.
	// tracer := opentracing.GlobalTracer()

	// todo 暂时内部，只支持 http, 后面补充https
	host := fmt.Sprintf("http://%s", url.Location)
	conn := http.DefaultClient

	key := url.GetParam(constant.InterfaceKey, "")
	impl := service.GetConsumerServiceByInterfaceName(key)
	logger.Infof("%v", url.String())

	invoker := getInvoker(impl, conn, host)
	return &Client{
		Client:  conn,
		invoker: reflect.ValueOf(invoker),
	}, nil
}

func getInvoker(impl interface{}, conn *http.Client, host string) interface{} {
	var in []reflect.Value
	in = append(in, reflect.ValueOf(conn), reflect.ValueOf(host))
	method := reflect.ValueOf(impl).MethodByName("GetGostStub")
	res := method.Call(in)
	return res[0].Interface()
}
