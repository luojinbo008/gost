import (
	"net/http"
	"github.com/labstack/echo/v4"
)

{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{.ServiceType}}Server interface {
	{{- range .Methods}}
	// Sends a greeting
	{{.Name}}(req *Echo{{.Request}}, e echo.Context) (*Echo{{.Reply}}, error) 
	{{- end}}
}

func Register{{.ServiceType}}Router(e *echo.Echo, s {{.ServiceType}}Server) {
	{{- range .Methods}}
	e.{{.Method}}("{{.Path}}", func (c echo.Context) error{
		var req *Echo{{.Request}} = new(Echo{{.Request}})
		if err := c.Bind(req); err != nil {
			return err
		}
		reply, err := s.{{.Name}}(req, c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
	{{- end}}
}