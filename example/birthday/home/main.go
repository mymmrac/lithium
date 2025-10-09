package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"

	"github.com/extism/go-pdk"

	plugin_protocol "github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:embed index.gohtml
var indexTemplate string

//go:wasmexport handler
func Handle() {
	var request plugin_protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	index, err := template.New("").Parse(indexTemplate)
	if err != nil {
		pdk.SetError(fmt.Errorf("parse template: %w", err))
		return
	}

	// Pass Google Client ID into template for GIS
	data := map[string]any{
		"ClientID": os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	}

	response := plugin_protocol.Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	}

	buf := bytes.NewBuffer(nil)
	if err = index.Execute(buf, data); err != nil {
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
