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
	// Tags for this watch (separate from tags on the watched item itself).
	Tags []Tag `json:"tags" gorm:"many2many:activity_details_tags;"`
}
