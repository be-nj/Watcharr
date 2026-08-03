package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sbondCo/Watcharr/database/entity"
)

// Refresh the cached OSM data of a cinema (name stays untouched, it
// may be customised). Explicit user action only - the cache is never
// synced automatically (ADR 0003). If the object is gone upstream we
// keep the cached data and just report it.
func (s *Service) RefreshCinemaOsm(sourceId uint) (entity.CinemaDetails, error) {
	source, err := s.GetSource(0, sourceId)
	if err != nil {
		return entity.CinemaDetails{}, err
	}
	if source.Type != entity.SOURCE_CINEMA || source.Cinema == nil {
		return entity.CinemaDetails{}, errors.New("source is not a cinema")
	}
	cinema := source.Cinema
	if cinema.OsmID == 0 {
		return entity.CinemaDetails{}, errors.New("cinema has no osm reference")
	}
	result, err := s.osmLookup(cinema.OsmType, cinema.OsmID)
	if err != nil {
		return entity.CinemaDetails{}, err
	}
	if result == nil {
		return entity.CinemaDetails{}, errors.New(
			"the osm object is gone upstream, keeping the cached data")
	}
	if lat, err := strconv.ParseFloat(result.Lat, 64); err == nil {
		cinema.Lat = &lat
	}
	if lon, err := strconv.ParseFloat(result.Lon, 64); err == nil {
		cinema.Lon = &lon
	}
	if city := firstNonEmpty(
		result.Address.City,
		result.Address.Town,
		result.Address.Village,
	); city != "" {
		cinema.City = city
	}
	if addr := strings.TrimSpace(fmt.Sprintf("%s %s",
		result.Address.Road, result.Address.HouseNumber)); addr != "" {
		cinema.Address = addr
	}
	if result.Extratags.Wikidata != "" {
		cinema.WikidataID = result.Extratags.Wikidata
	}
	if cinema.WikidataID != "" {
		if website, err := s.wikidataWebsite(cinema.WikidataID); err != nil {
			// Enrichment only - a failure shouldn't fail the refresh.
			slog.Error("refreshCinemaOsm: Wikidata enrichment failed",
				"wikidata_id", cinema.WikidataID, "error", err)
		} else if website != "" {
			cinema.Website = website
		}
	}
	res := s.db.Save(cinema)
	if res.Error != nil {
		slog.Error("refreshCinemaOsm: Failed saving refreshed details",
			"source_id", sourceId, "error", res.Error.Error())
		return entity.CinemaDetails{}, errors.New("failed saving refreshed details")
	}
	return *cinema, nil
}

// Look up a single object by its OSM reference.
// https://nominatim.org/release-docs/latest/api/Lookup/
// Returns nil (no error) when the object no longer exists.
func (s *Service) osmLookup(osmType string, osmID int64) (*GeocodeResult, error) {
	prefix := map[string]string{"node": "N", "way": "W", "relation": "R"}[osmType]
	if prefix == "" {
		return nil, errors.New("invalid osm type")
	}
	base, err := url.Parse("https://nominatim.openstreetmap.org/lookup")
	if err != nil {
		return nil, errors.New("failed to parse lookup api uri")
	}
	params := url.Values{}
	params.Add("osm_ids", fmt.Sprintf("%s%d", prefix, osmID))
	params.Add("format", "json")
	params.Add("addressdetails", "1")
	params.Add("extratags", "1")
	base.RawQuery = params.Encode()
	body, err := nominatimGet(base)
	if err != nil {
		return nil, err
	}
	results := new([]GeocodeResult)
	if err := json.Unmarshal(body, results); err != nil {
		return nil, err
	}
	if len(*results) == 0 {
		return nil, nil
	}
	return &(*results)[0], nil
}

// Official website (P856) of a wikidata entity, empty when not set.
// https://www.wikidata.org/wiki/Special:EntityData
func (s *Service) wikidataWebsite(wikidataId string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(
		"https://www.wikidata.org/wiki/Special:EntityData/%s.json",
		url.PathEscape(wikidataId)), nil)
	if err != nil {
		return "", err
	}
	// Wikimedia rejects requests without an identifying user agent.
	req.Header.Add("User-Agent", "watcharr-fork/1.0 (https://github.com/be-nj/Watcharr)")
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", fmt.Errorf("wikidata status code %d", res.StatusCode)
	}
	var data struct {
		Entities map[string]struct {
			Claims map[string][]struct {
				Mainsnak struct {
					Datavalue struct {
						Value json.RawMessage `json:"value"`
					} `json:"datavalue"`
				} `json:"mainsnak"`
			} `json:"claims"`
		} `json:"entities"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	for _, entity := range data.Entities {
		if sites, ok := entity.Claims["P856"]; ok && len(sites) > 0 {
			var website string
			if err := json.Unmarshal(
				sites[0].Mainsnak.Datavalue.Value, &website); err == nil {
				return website, nil
			}
		}
	}
	return "", nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
