package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/extism/go-pdk"

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
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		respond(protocol.Response{
			StatusCode: http.StatusBadRequest,
			Body:       "failed to parse request body: " + err.Error(),
		})
		return
	}

	location, err := getLocationInfo(body.Location)
	if err != nil {
		respond(protocol.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       "failed to get coordinates: " + err.Error(),
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
		"name":      location.Name,
		"country":   location.Country,
		"latitude":  location.Latitude,
		"longitude": location.Longitude,
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

type geoResponse struct {
	Results []getResult `json:"results"`
}

type getResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
}

func getLocationInfo(location string) (*getResult, error) {
	geoURL := "https://geocoding-api.open-meteo.com/v1/search"
	params := url.Values{}
	params.Set("name", location)
	params.Set("count", "1")

	resp, err := http.Get(geoURL + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding returned status %d", resp.StatusCode)
	}

	var geoRes geoResponse
	if err = json.NewDecoder(resp.Body).Decode(&geoRes); err != nil {
		return nil, fmt.Errorf("decode geocoding response failed: %w", err)
	}
	if len(geoRes.Results) == 0 {
		return nil, errors.New("no results found for location")
	}

	return &geoRes.Results[0], nil
}

func main() {}
