package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/extism/go-pdk"

	"github.com/mymmrac/lithium/pkg/plugin/network"
	"github.com/mymmrac/lithium/pkg/plugin/protocol"
)

//go:wasmexport handler
func Handle() {
	network.PatchDefaultHTTPClient()

	var request protocol.Request
	if err := pdk.InputJSON(&request); err != nil {
		pdk.SetError(fmt.Errorf("unmarshal request: %w", err))
		return
	}

	var body struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		respond(protocol.Response{
			StatusCode: http.StatusBadRequest,
			Body:       "failed to parse request body: " + err.Error(),
		})
		return
	}

	temperature, perception, err := getWeatherInfo(body.Latitude, body.Longitude)
	if err != nil {
		respond(protocol.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       "failed to get weather: " + err.Error(),
		})
		return
	}

	response := protocol.Response{
		StatusCode: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}

	responseBody, err := json.Marshal(map[string]any{
		"temperature": temperature,
		"perception":  perception,
	})
	if err != nil {
		respond(protocol.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       "marshal response: " + err.Error(),
		})
		return
	}
	response.Body = string(responseBody)

	respond(response)
}

func respond(response protocol.Response) {
	if err := pdk.OutputJSON(response); err != nil {
		pdk.SetError(fmt.Errorf("marshal response: %w", err))
		return
	}
}

type forecastResponse struct {
	Current struct {
		Temperature              float64 `json:"temperature_2m"`
		PrecipitationProbability float64 `json:"precipitation_probability"`
	} `json:"current"`
}

func getWeatherInfo(lat, lon float64) (float64, float64, error) {
	weatherURL := "https://api.open-meteo.com/v1/forecast"
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", lat))
	params.Set("longitude", fmt.Sprintf("%f", lon))
	params.Set("current", "temperature_2m,precipitation_probability")
	params.Set("timezone", "auto")

	resp, err := http.Get(weatherURL + "?" + params.Encode())
	if err != nil {
		return 0, 0, fmt.Errorf("weather request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var wRes forecastResponse
	if err = json.NewDecoder(resp.Body).Decode(&wRes); err != nil {
		return 0, 0, fmt.Errorf("decode forecast response failed: %w", err)
	}

	return wRes.Current.Temperature, wRes.Current.PrecipitationProbability, nil
}

func main() {}
