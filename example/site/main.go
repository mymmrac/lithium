//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o site.wasm main.go
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/extism/go-pdk"
)

//go:embed index.gohtml
var indexTemplate string

//go:wasmexport handler
func Handle() {
	var request Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	index, err := template.New("").Parse(indexTemplate)
	if err != nil {
		pdk.SetError(fmt.Errorf("parse template: %w", err))
		return
	}

	response := Response{
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

type Request struct {
	URL     string              `json:"url"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type Response struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

func main() {}
