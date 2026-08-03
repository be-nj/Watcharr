package domain

type (
	WatchSourceAddRequest struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required"`
		// Cinema fields (osm reference + cached master data), only
		// used when type is CINEMA. Empty osm fields = free form cinema.
		City       string   `json:"city"`
		Address    string   `json:"address"`
		Lat        *float64 `json:"lat"`
		Lon        *float64 `json:"lon"`
		OsmType    string   `json:"osmType"`
		OsmID      int64    `json:"osmId"`
		WikidataID string   `json:"wikidataId"`
	}

	CinemaDetailsUpdateRequest struct {
		City    string   `json:"city"`
		Address string   `json:"address"`
		Lat     *float64 `json:"lat"`
		Lon     *float64 `json:"lon"`
		Note    string   `json:"note"`
	}

	CinemaScreenAddRequest struct {
		Name string `json:"name" binding:"required"`
	}

	// A single visit rating of a source (for the source detail page).
	// Username is empty unless the rater opted into showing it.
	WatchSourceRating struct {
		Date          string   `json:"date"`
		RatingOverall *float64 `json:"ratingOverall"`
		RatingSnacks  *float64 `json:"ratingSnacks"`
		RatingTech    *float64 `json:"ratingTech"`
		RatingComfort *float64 `json:"ratingComfort"`
		Username      string   `json:"username"`
		Own           bool     `json:"own"`
	}

	// A single watch that used a source (for the source detail page).
	WatchSourceWatch struct {
		ActivityID uint   `json:"activityId"`
		WatchedID  uint   `json:"watchedId"`
		Date       string `json:"date"`
		ScreenName string `json:"screenName"`
		Title      string `json:"title"`
		Type       string `json:"type"`
		TmdbID     int    `json:"tmdbId"`
		PosterPath string `json:"poster_path"`
	}
)
