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
// streaming service, live tv, ...). Sources belong to the instance and
// are shared by all users; CreatedBy is provenance, not ownership.
// Only cinema sources carry extra details.
type WatchSource struct {
	dbmodel.GormModel
	// ID of user that created this source.
	CreatedBy uint `json:"createdBy" gorm:"not null"`
	// Display name of the source (eg `CineStar Metropolis` or `Disney+`).
	Name string          `json:"name" gorm:"not null"`
	Type WatchSourceType `json:"type" gorm:"not null"`
	// Only set for sources of type cinema.
	Cinema *CinemaDetails `json:"cinema,omitempty" gorm:"foreignKey:WatchSourceID"`
	// Visit rating aggregate (not stored, filled when listing sources).
	RatingAverage *float64 `json:"ratingAverage,omitempty" gorm:"-"`
	RatingCount   int64    `json:"ratingCount,omitempty" gorm:"-"`
}

type CinemaDetails struct {
	dbmodel.GormModelNoDel
	WatchSourceID uint     `json:"-" gorm:"uniqueIndex;not null"`
	City          string   `json:"city"`
	Address       string   `json:"address"`
	Lat           *float64 `json:"lat"`
	Lon           *float64 `json:"lon"`
	// Reference to the real world cinema on OpenStreetMap. The local
	// name/coordinates/address are a CACHE of it: never auto-synced,
	// refreshed only on explicit user action, and never cleared when
	// the object disappears upstream. Empty for cinemas without an OSM
	// entry (free form fallback).
	OsmType    string `json:"osmType"`
	OsmID      int64  `json:"osmId"`
	WikidataID string `json:"wikidataId"`
	// Free text notes (no-gos like `has no popcorn!`), shared by the
	// instance like the cinema itself.
	Note string `json:"note"`
	// Screens of this cinema (a watch can reference which screen it was in).
	Screens []CinemaScreen `json:"screens,omitempty" gorm:"foreignKey:CinemaDetailsID"`
}

type CinemaScreen struct {
	dbmodel.GormModelNoDel
	CinemaDetailsID uint   `json:"-" gorm:"not null"`
	Name            string `json:"name" gorm:"not null"`
}
