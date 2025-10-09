package main

import (
	"fmt"
	"net/http"

	"github.com/extism/go-pdk"

	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	// Expire session cookie
	err := pdk.OutputJSON(protocol.Response{
		StatusCode: http.StatusNoContent,
		Headers: map[string][]string{
			"Set-Cookie": {"session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax"},
		},
	})
	if err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
