package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/extism/go-pdk"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mymmrac/lithium/pkg/plugin/network"
	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	if err := network.PatchDefaultHTTPClient(); err != nil {
		pdk.SetError(fmt.Errorf("pathc http client: %w", err))
		return
	}

	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil || body.IDToken == "" {
		respond(http.StatusBadRequest, nil, "invalid body")
		return
	}

	response, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + body.IDToken)
	if err != nil || response.StatusCode != http.StatusOK {
		respond(http.StatusUnauthorized, nil, "invalid token")
		return
	}
	defer func() { _ = response.Body.Close() }()

	var tokenInfo struct {
		Aud   string `json:"aud"`
		Email string `json:"email"`
	}
	if err = json.NewDecoder(response.Body).Decode(&tokenInfo); err != nil || tokenInfo.Email == "" {
		respond(http.StatusUnauthorized, nil, "decode auth response")
		return
	}

	if tokenInfo.Aud != os.Getenv("GOOGLE_OAUTH_CLIENT_ID") {
		respond(http.StatusUnauthorized, nil, "aud mismatch")
		return
	}

	ttl := time.Hour
	token, err := generateJWT(tokenInfo.Email, time.Now().Add(ttl))
	if err != nil {
		respond(http.StatusInternalServerError, nil, "generate jwt token")
		return
	}

	cookie := fmt.Sprintf("session=%s; Path=/; HttpOnly; SameSite=Lax; Max-Age=%d", token, int(ttl.Seconds()))
	respond(http.StatusNoContent, map[string][]string{"Set-Cookie": {cookie}}, "")
}

func generateJWT(email string, expiresAt time.Time) (string, error) {
	now := jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: now,
		IssuedAt:  now,
	})

	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func respond(status int, headers map[string][]string, body string) {
	if headers == nil {
		headers = map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}}
	}
	err := pdk.OutputJSON(protocol.Response{StatusCode: status, Headers: headers, Body: body})
	if err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

func main() {}
