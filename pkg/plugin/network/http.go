//go:build wasip1

package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/extism/go-pdk"
	"github.com/mymmrac/wape/plugin/net"
)

// TODO: Return error
func HTTPClient() *http.Client {
	// TODO: Read from env
	caPEM, err := os.ReadFile("/certs/ca-certificates.crt")
	if err != nil {
		pdk.SetError(fmt.Errorf("reading CA cert: %w", err))
		return nil
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caPEM); !ok {
		pdk.SetError(fmt.Errorf("unable to append CA cert"))
		return nil
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext:     net.DefaultDialer.DialContext,
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
	}
}

func PatchDefaultHTTPClient() {
	http.DefaultClient = HTTPClient()
}
