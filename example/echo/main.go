//go:generate env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -ldflags "-s -w" -o echo.wasm main.go
package main

import (
	"fmt"
	"strconv"

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
		Body:       "",
	}

	response.Body += "URL: " + strconv.Quote(request.URL) + "\n"
	response.Body += "Method: " + strconv.Quote(request.Method) + "\n"
	response.Body += "Headers:\n"
	for key, values := range request.Headers {
		response.Body += "\t" + strconv.Quote(key) + ":"
		for i, value := range values {
			response.Body += strconv.Quote(value)
			if i < len(values)-1 {
				response.Body += ", "
			}
		}
		response.Body += "\n"
	}
	response.Body += "Body:\n"
	response.Body += strconv.Quote(request.Body)

	if err := pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
