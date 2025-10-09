package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/extism/go-pdk"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	token := readCookie(request.Headers, "session")
	if token == "" {
		respondJSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var claims jwt.RegisteredClaims
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		respondJSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if !parsedToken.Valid {
		respondJSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	respondJSON(http.StatusOK, map[string]string{"email": claims.Subject})
}

func readCookie(headers map[string][]string, name string) string {
	for k, vs := range headers {
		if strings.EqualFold(k, "Cookie") && len(vs) > 0 {
			parts := strings.Split(vs[0], ";")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if strings.HasPrefix(p, name+"=") {
					return strings.TrimPrefix(p, name+"=")
				}
			}
		}
	}
	return ""
}

func respondJSON(code int, body any) {
	data, _ := json.Marshal(body)
	err := pdk.OutputJSON(protocol.Response{
		StatusCode: code,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       string(data),
	})
	if err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
