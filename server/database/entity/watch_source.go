package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

type WatchSourceType string

const (
	SOURCE_CINEMA     WatchSourceType = "CINEMA"
	SOURCE_STREAMING  WatchSourceType = "STREAMING"
	SOURCE_TV         WatchSourceType = "TV"
	SOURCE_SELFHOSTED WatchSourceType = "SELFHOSTED"
	SOURCE_DISC       WatchSourceType = "DISC"
	SOURCE_OTHER      WatchSourceType = "OTHER"
)

func (t WatchSourceType) IsValid() bool {
	switch t {
	case SOURCE_CINEMA,
		SOURCE_STREAMING,
		SOURCE_TV,
		SOURCE_SELFHOSTED,
		SOURCE_DISC,
		SOURCE_OTHER:
		return true
	}
	return false
}

// A watch source describes where/how a watch happened (a cinema, a
// streaming service, live tv, ...). Owned by a user, referenced by their
// activities. Only cinema sources carry extra details.
type WatchSource struct {
	dbmodel.GormModel
	// ID of user that owns this source.
	UserID uint `json:"-" gorm:"not null"`
	// Display name of the source (eg `CineStar Metropolis` or `Disney+`).
	Name string          `json:"name" gorm:"not null"`
	Type WatchSourceType `json:"type" gorm:"not null"`
	// Only set for sources of type cinema.
	Cinema *CinemaDetails `json:"cinema,omitempty" gorm:"foreignKey:WatchSourceID"`
}

type CinemaDetails struct {
	dbmodel.GormModelNoDel
	WatchSourceID uint     `json:"-" gorm:"uniqueIndex;not null"`
	City          string   `json:"city"`
	Address       string   `json:"address"`
	Lat           *float64 `json:"lat"`
	Lon           *float64 `json:"lon"`
	// Free text notes (no-gos like `has no popcorn!`).
	Note string `json:"note"`
	// Optional ratings, every dimension can be left empty.
	// Like watched ratings these are always saved as out of 10.0.
	RatingOverall *float64 `json:"ratingOverall" gorm:"type:numeric(2,1)"`
	RatingSnacks  *float64 `json:"ratingSnacks" gorm:"type:numeric(2,1)"`
	RatingTech    *float64 `json:"ratingTech" gorm:"type:numeric(2,1)"`
	RatingComfort *float64 `json:"ratingComfort" gorm:"type:numeric(2,1)"`
	// Screens of this cinema (a watch can reference which screen it was in).
	Screens []CinemaScreen `json:"screens,omitempty" gorm:"foreignKey:CinemaDetailsID"`
}

type CinemaScreen struct {
	dbmodel.GormModelNoDel
	CinemaDetailsID uint   `json:"-" gorm:"not null"`
	Name            string `json:"name" gorm:"not null"`
}
