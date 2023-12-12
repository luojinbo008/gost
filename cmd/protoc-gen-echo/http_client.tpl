import (
	"io/ioutil"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"fmt"
	
	"github.com/luojinbo008/gost/cmd/runtime"
)

{{$svrType := .ServiceType}}
type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{.Name}}(ctx context.Context, req *Echo{{.Request}}) (rsp *Echo{{.Reply}}, err error)
{{- end}}
}

type {{.ServiceType}}HTTPClientImpl struct{
	cc *http.Client
	host string
}

func New{{.ServiceType}}HTTPClient (client *http.Client, host string) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{client, host}
}

// Reference ...
func (c *{{$svrType}}HTTPClientImpl) Reference() string {
	return "{{$svrType}}HTTPClientImpl"
}


{{range .MethodSets}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *Echo{{.Request}}) (*Echo{{.Reply}}, error) {
	var out {{.Reply}}
	pattern := runtime.Path(fmt.Sprintf("%s%s", c.host, "{{.Path}}"))

	res, err := c.Call(ctx, "GET", pattern, in)
	if err != nil {
		return &out, err
	}

	err = json.Unmarshal(res, &out)
	if err != nil {
		return &out, err
	}
	return &out, err
}
{{end}}

func (c *HelloworldHTTPClientImpl) GetGostStub(cc *http.Client, host string) HelloworldHTTPClient {
	return NewHelloworldHTTPClient(cc, host)
}

func (c *HelloworldHTTPClientImpl) Call(ctx context.Context, method string, pattern runtime.Path, in interface{}) ([]byte, error) {
	body, err := runtime.Values(pattern, in)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(body.GetPath())
	if err != nil {
		return nil, err
	}

	if body.GetQuery() != nil {
		query := u.Query()
		u.RawQuery = query.Encode()
	}

	reqURL := u.String()
	var paramBody []byte
	paramBody, err = json.Marshal(in)
	if err != nil {
		return nil, err
	}

	// Request instant
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewBuffer(paramBody))
	if err != nil {
		return nil, err
	}

	// headers
	req.Header = body.GetHeader()
	req.Header.Add("Content-type", "application/json")

	// do request
	resp, err := c.cc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return res, nil
}