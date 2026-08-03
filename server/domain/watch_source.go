package domain

type (
	WatchSourceAddRequest struct {
		Name string `json:"name" binding:"required"`
		Type string `json:"type" binding:"required"`
	}

	CinemaDetailsUpdateRequest struct {
		City          string   `json:"city"`
		Address       string   `json:"address"`
		Lat           *float64 `json:"lat"`
		Lon           *float64 `json:"lon"`
		Note          string   `json:"note"`
		RatingOverall *float64 `json:"ratingOverall" binding:"omitempty,max=10"`
		RatingSnacks  *float64 `json:"ratingSnacks" binding:"omitempty,max=10"`
		RatingTech    *float64 `json:"ratingTech" binding:"omitempty,max=10"`
		RatingComfort *float64 `json:"ratingComfort" binding:"omitempty,max=10"`
	}

	CinemaScreenAddRequest struct {
		Name string `json:"name" binding:"required"`
	}
)
