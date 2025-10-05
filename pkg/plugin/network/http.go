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

func TLSConfig() (*tls.Config, error) {
	caPEM, err := os.ReadFile(os.Getenv("LITHIUM_CA_CERT_FILE"))
	if err != nil {
		return nil, fmt.Errorf("reading CA cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("unable to append CA cert")
	}

	return &tls.Config{
		RootCAs: certPool,
	}, nil
}

func HTTPClient() (*http.Client, error) {
	tlsConfig, err := TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("create TLS config: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext:     net.DefaultDialer.DialContext,
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

func PatchDefaultHTTPClient() error {
	client, err := HTTPClient()
	if err != nil {
		return fmt.Errorf("create HTTP client: %w", err)
	}

	http.DefaultClient = client
	return nil
}
