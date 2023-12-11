package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed http_router.tpl
var routerTemplate string

//go:embed http_server.tpl
var serverTemplate string

//go:embed http_client.tpl
var clientTemplate string

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {

	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	Comment      string

	// http_rule
	Path         string
	Method       string
	HasVars      bool
	HasBody      bool
	Body         string
	ResponseBody string

	Fields []*RequestField
}

type RequestField struct {
	ProtoName string
	GoName    string
	GoType    string
	ConvExpr  string
}

func (s *serviceDesc) execute(tpl string) string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	var tmpl *template.Template
	var err error

	switch tpl {
	case "router":
		tmpl, err = template.New("http").Parse(strings.TrimSpace(routerTemplate))
	case "server":
		tmpl, err = template.New("http").Parse(strings.TrimSpace(serverTemplate))
	case "client":
		tmpl, err = template.New("http").Parse(strings.TrimSpace(clientTemplate))
	default:
		panic("tpl not found")
	}

	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
