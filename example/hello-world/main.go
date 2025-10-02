//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o hello-world.wasm main.go
package main

import (
	"fmt"

	"github.com/extism/go-pdk"
)

//go:wasmexport handler
func Handle() {
	var request Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	response := Response{
		StatusCode: 200,
		Headers:    nil,
		Body:       "Hello World!",
	}

	if err := pdk.OutputJSON(response); err != nil {
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
