//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o hello-world.wasm main.go
package main

import (
	"fmt"

	"github.com/extism/go-pdk"

	"github.com/mymmrac/lithium/pkg/module/protocol"
)

//go:wasmexport handler
func Handle() {
	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	response := protocol.Response{
		StatusCode: 200,
		Headers:    nil,
		Body:       "Hello World!",
	}

	if err := pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
