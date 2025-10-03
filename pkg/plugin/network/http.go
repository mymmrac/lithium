//go:build wasip1

package network

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/mymmrac/wape/plugin/net"
)

func HTTPClient() (*http.Client, error) {
	caPEM, err := os.ReadFile(os.Getenv("LITHIUM_CA_CERT_FILE"))
	if err != nil {
		return nil, fmt.Errorf("reading CA cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("unable to append CA cert")
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext:     net.DefaultDialer.DialContext,
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
	}, nil
}

func PatchDefaultHTTPClient() error {
	client, err := HTTPClient()
	if err != nil {
		return err
	}

	http.DefaultClient = client
	return nil
}
