package source

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Nominatims usage policy allows at most 1 request/second and requires
// an identifying user agent - throttle all geocode calls process wide.
var (
	geocodeMu       sync.Mutex
	geocodeLastCall time.Time
)

type GeocodeResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	// OSM reference + object kind, for anchoring cinemas (ADR 0003).
	OsmType string `json:"osm_type"`
	OsmID   int64  `json:"osm_id"`
	Class   string `json:"class"`
	Type    string `json:"type"`
	Address struct {
		City        string `json:"city"`
		Town        string `json:"town"`
		Village     string `json:"village"`
		Road        string `json:"road"`
		HouseNumber string `json:"house_number"`
	} `json:"address"`
	Extratags struct {
		Wikidata string `json:"wikidata"`
	} `json:"extratags"`
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
	params.Add("addressdetails", "1")
	params.Add("extratags", "1")
	base.RawQuery = params.Encode()
	body, err := nominatimGet(base)
	if err != nil {
		return nil, err
	}
	results := new([]GeocodeResult)
	err = json.Unmarshal(body, results)
	if err != nil {
		return nil, err
	}
	return *results, nil
}

// Shared nominatim http call: throttled process wide, with the
// identifying user agent their usage policy requires.
func nominatimGet(base *url.URL) ([]byte, error) {
	geocodeMu.Lock()
	if wait := 1100*time.Millisecond - time.Since(geocodeLastCall); wait > 0 {
		time.Sleep(wait)
	}
	geocodeLastCall = time.Now()
	geocodeMu.Unlock()

	client := &http.Client{}
	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		slog.Error("geocode: Creating request to nominatim failed", "error", err)
		return nil, errors.New("request failed")
	}
	// Nominatim usage policy requires an identifying user agent.
	req.Header.Add("User-Agent", "watcharr-fork/1.0 (https://github.com/be-nj/Watcharr)")
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
	if res.StatusCode == 429 {
		slog.Error("geocode: Nominatim rate limited us", "status_code", res.StatusCode)
		return nil, errors.New("openstreetmap search is rate limited right now, try again in a minute")
	}
	if res.StatusCode != 200 {
		slog.Error("geocode: Nominatim non 200 status code", "status_code", res.StatusCode, "error", string(body))
		return nil, errors.New("geocoding failed")
	}
	return body, nil
}
