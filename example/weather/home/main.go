package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/extism/go-pdk"

	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:embed index.gohtml
var indexTemplate string

//go:wasmexport handler
func Handle() {
	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	index, err := template.New("").Parse(indexTemplate)
	if err != nil {
		pdk.SetError(fmt.Errorf("parse template: %w", err))
		return
	}

	response := protocol.Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	}

	buf := bytes.NewBuffer(nil)
	if err = index.Execute(buf, request); err != nil {
		pdk.SetError(fmt.Errorf("execute template: %w", err))
		return
	}

	response.Body = buf.String()

	if err = pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
