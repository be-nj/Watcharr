package domain

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

type (
	// Internal struct accepted by AddActivity function.
	ActivityAddProps struct {
		WatchedID  uint                `json:"watchedId" binding:"required"`
		Type       entity.ActivityType `json:"type" binding:"required"`
		Data       string              `json:"data" binding:"required"`
		CustomDate *time.Time          `json:"customDate,omitempty"`
	}

	ActivityUpdateRequest struct {
		CustomDate time.Time `json:"customDate" binding:"required"`
	}

	// Log one of our plays for friends that watched together with us.
	CompanionLogRequest struct {
		UserIDs []uint `json:"userIds" binding:"required"`
	}

	CompanionLogResult struct {
		UserID   uint   `json:"userId"`
		Username string `json:"username,omitempty"`
		Ok       bool   `json:"ok"`
		Error    string `json:"error,omitempty"`
	}

	ActivityDetailsUpdateRequest struct {
		WatchSourceID  *uint  `json:"watchSourceId"`
		CinemaScreenID *uint  `json:"cinemaScreenId"`
		AudioLang      string `json:"audioLang"`
		SubtitleLang   string `json:"subtitleLang"`
		Note           string `json:"note"`
		// Visit rating for the cinema of this watch. Overall is
		// required as soon as any dimension is set.
		RatingOverall  *float64 `json:"ratingOverall" binding:"omitempty,max=10"`
		RatingSnacks   *float64 `json:"ratingSnacks" binding:"omitempty,max=10"`
		RatingTech     *float64 `json:"ratingTech" binding:"omitempty,max=10"`
		RatingComfort  *float64 `json:"ratingComfort" binding:"omitempty,max=10"`
		RatingShowName bool     `json:"ratingShowName"`
	}

	ActivityAddProvider interface {
		AddActivity(
			userId uint,
			ar ActivityAddProps,
			countAsPlay bool,
		) (entity.Activity, error)
	}
)

// Looks through Activity for Watched entry and calculates the amount
// that count as plays.
func getPlaysFromActivity(a []entity.Activity) int {
	plays := 0
	for i := range a {
		if a[i].CountAsPlay {
			plays++
		}
	}
	return plays
}
