package source

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type GeocodeResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

// Geocode a free text query (eg cinema name + city) via the public
// nominatim api, so coordinates can be prefilled instead of typed.
// https://nominatim.org/release-docs/latest/api/Search/
func (s *Service) Geocode(query string) ([]GeocodeResult, error) {
	base, err := url.Parse("https://nominatim.openstreetmap.org/search")
	if err != nil {
		return nil, errors.New("failed to parse geocode api uri")
	}
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("limit", "5")
	base.RawQuery = params.Encode()

	client := &http.Client{}
	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		slog.Error("geocode: Creating request to nominatim failed", "error", err)
		return nil, errors.New("request failed")
	}
	// Nominatim usage policy requires an identifying user agent.
	req.Header.Add("User-Agent", "Watcharr")
	res, err := client.Do(req)
	if err != nil {
		slog.Error("geocode: Making request to nominatim failed", "error", err)
		return nil, errors.New("request failed")
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error("geocode: Error reading nominatim response", "error", err.Error())
		return nil, err
	}
	if res.StatusCode != 200 {
		slog.Error("geocode: Nominatim non 200 status code", "status_code", res.StatusCode, "error", string(body))
		return nil, errors.New("geocoding failed")
	}
	results := new([]GeocodeResult)
	err = json.Unmarshal(body, results)
	if err != nil {
		return nil, err
	}
	return *results, nil
}
