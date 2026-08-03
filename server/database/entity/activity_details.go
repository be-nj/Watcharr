package entity

import "github.com/sbondCo/Watcharr/database/dbmodel"

// Extra metadata a user can attach to a single watch (an activity):
// where/how it was watched, language/subtitles, tags and a note.
// Created lazily, so most activities won't have a row here.
type ActivityDetails struct {
	dbmodel.GormModelNoDel
	ActivityID uint `json:"-" gorm:"uniqueIndex;not null"`
	// Where/how this watch happened (one of the users watch sources).
	WatchSourceID *uint        `json:"watchSourceId"`
	WatchSource   *WatchSource `json:"watchSource,omitempty" gorm:"foreignKey:WatchSourceID"`
	// Which screen it was in (only for cinema sources).
	CinemaScreenID *uint         `json:"cinemaScreenId"`
	CinemaScreen   *CinemaScreen `json:"cinemaScreen,omitempty" gorm:"foreignKey:CinemaScreenID"`
	// Audio language (lowercase iso 639-1 code, eg `de`). Empty when not set.
	AudioLang string `json:"audioLang"`
	// Subtitle language (lowercase iso 639-1 code), `none` when subtitles
	// were explicitly off. Empty when not set.
	SubtitleLang string `json:"subtitleLang"`
	// Free text note for this watch.
	Note string `json:"note"`
	// Visit rating for the cinema this watch happened in (only valid
	// with a cinema watch source). All dimensions optional, but overall
	// is required as soon as any dimension is rated (validated in the
	// service). Like watched ratings these are always out of 10.0.
	RatingOverall *float64 `json:"ratingOverall" gorm:"type:numeric(2,1)"`
	RatingSnacks  *float64 `json:"ratingSnacks" gorm:"type:numeric(2,1)"`
	RatingTech    *float64 `json:"ratingTech" gorm:"type:numeric(2,1)"`
	RatingComfort *float64 `json:"ratingComfort" gorm:"type:numeric(2,1)"`
	// Whether the users name may be shown next to this rating (opt in,
	// default anonymous).
	RatingShowName bool `json:"ratingShowName"`
}
