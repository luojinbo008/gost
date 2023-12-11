import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
)

{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{range .Methods}}
func {{$svrType}}{{.Name}}BusinessHandler(req *Echo{{.Request}}, c echo.Context) (Echo{{.Reply}}, error) {
	rj, err := json.Marshal(req)
	if err != nil {
		return Echo{{.Reply}}{}, err
	}
	fmt.Printf("Got {{.Request}} is: %v\n", string(rj))
	return Echo{{.Reply}}{}, nil 
}
{{end}}