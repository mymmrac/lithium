package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/extism/go-pdk"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mymmrac/lithium/pkg/plugin/network"
	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	if err := network.PatchDefaultHTTPClient(); err != nil {
		pdk.SetError(fmt.Errorf("patch http client: %w", err))
		return
	}

	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	token := readCookie(request.Headers, "session")
	if token == "" {
		respond(protocol.Response{StatusCode: http.StatusUnauthorized})
		return
	}

	var claims jwt.RegisteredClaims
	parsedToken, err := jwt.ParseWithClaims(token, &claims, func(_ *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		respond(protocol.Response{StatusCode: http.StatusUnauthorized})
		return
	}

	if !parsedToken.Valid {
		respond(protocol.Response{StatusCode: http.StatusUnauthorized})
		return
	}

	var body struct {
		Recipient string `json:"recipient"`
		Age       *int   `json:"age"`
		Tone      string `json:"tone"`
		Details   string `json:"details"`
	}
	if err = json.Unmarshal([]byte(request.Body), &body); err != nil {
		respond(protocol.Response{StatusCode: http.StatusBadRequest, Body: "invalid body"})
		return
	}
	if body.Recipient == "" {
		respond(protocol.Response{StatusCode: http.StatusBadRequest, Body: "recipient required"})
		return
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		respond(protocol.Response{StatusCode: http.StatusInternalServerError, Body: "groq api key not configured"})
		return
	}

	message, err := generateBirthdayMessage(apiKey, body.Recipient, body.Age, body.Tone, body.Details)
	if err != nil {
		respond(protocol.Response{StatusCode: http.StatusInternalServerError, Body: "generation failed: " + err.Error()})
		return
	}

	resp := map[string]any{"message": message}
	respBytes, _ := json.Marshal(resp)
	respond(protocol.Response{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       string(respBytes),
	})
}

func respond(response protocol.Response) {
	if err := pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
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

func generateBirthdayMessage(apiKey, recipient string, age *int, tone, details string) (string, error) {
	prompt := fmt.Sprintf("Compose a 'Happy Birthday' message for %s.", recipient)
	if age != nil {
		prompt += fmt.Sprintf(" They are %d years old.", *age)
	}
	if tone != "" {
		prompt += fmt.Sprintf(" Use a %s tone.", tone)
	}
	if details != "" {
		prompt += fmt.Sprintf(" Include these details: %s.", details)
	}
	prompt += " Keep it concise and warm. Respond in Ukrainian language."

	body := map[string]any{
		"model": os.Getenv("GROQ_MODEL"),
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are personal assistant for composing 'Happy Birthday' messages.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature":           1,
		"max_completion_tokens": 1024,
		"top_p":                 1,
		"stream":                false,
	}
	if body["model"] == "" {
		body["model"] = "llama-3.3-70b-versatile"
	}
	data, _ := json.Marshal(body)

	url := "https://api.groq.com/openai/v1/chat/completions"
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("grok api status: %d", response.StatusCode)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err = json.NewDecoder(response.Body).Decode(&out); err != nil {
		return "", err
	}

	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return out.Choices[0].Message.Content, nil
}

func main() {}
