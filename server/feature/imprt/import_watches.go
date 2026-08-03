package imprt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
)

// Merge watches (eg letterboxd diary entries) into an already existing
// watched entry:
//
//   - If the entry has exactly one play, it is replaced by the imported
//     watches (their dates are considered more trustworthy, eg a curated
//     letterboxd diary vs a bulk import timestamp).
//   - With multiple existing plays we merge conservatively: watches on
//     the same day only enrich the existing play with their metadata,
//     watches on new days are added as new plays.
//   - The rating is only set when none exists yet.
func (s *Service) mergeImportWatches(
	userId uint,
	ar *domain.ImportRequest,
	props domain.SuccessfulImportProps,
) domain.ImportResponse {
	var contentType entity.ContentType
	switch props.ContentType {
	case util.SupportedMediaMovie:
		contentType = entity.MOVIE
	case util.SupportedMediaShow:
		contentType = entity.SHOW
	default:
		slog.Error("mergeImportWatches: Unsupported content type.",
			"content_type", props.ContentType)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}
	}
	w, err := s.wp.GetWatchedItemByTmdbId(userId, uint(props.TmdbID), contentType)
	if err != nil {
		slog.Error("mergeImportWatches: Failed to get existing watched item.",
			"tmdb_id", props.TmdbID, "error", err)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}
	}
	// Only set the rating when none exists yet.
	if w.Rating == 0 && ar.Rating > 0 {
		res := s.db.
			Model(&entity.Watched{}).
			Where("id = ?", w.ID).
			Update("rating", ar.Rating)
		if res.Error != nil {
			slog.Error("mergeImportWatches: Failed to update rating.",
				"error", res.Error.Error())
		} else {
			w.Rating = ar.Rating
		}
	}
	if err := s.applyImportWatches(userId, w.ID, ar.Watches, true); err != nil {
		slog.Error("mergeImportWatches: Failed to apply watches.", "error", err)
		return domain.ImportResponse{Type: domain.IMPORT_FAILED}
	}
	return domain.ImportResponse{Type: domain.IMPORT_SUCCESS, WatchedEntry: w}
}

// Apply imported watches to a watched entry: watches on days that
// already have a play only get their metadata attached to that play,
// watches on new days are added as new play activities. When
// `replaceSinglePlay` is set and the entry has exactly one play, that
// play is removed first (imported watch dates win over it).
func (s *Service) applyImportWatches(
	userId uint,
	watchedId uint,
	watches []domain.ImportWatch,
	replaceSinglePlay bool,
) error {
	if len(watches) == 0 {
		return nil
	}
	activities := new([]entity.Activity)
	res := s.db.
		Model(&entity.Activity{}).
		Where("user_id = ? AND watched_id = ?", userId, watchedId).
		Find(&activities)
	if res.Error != nil {
		return errors.New("failed getting existing activities")
	}
	plays := []entity.Activity{}
	for _, a := range *activities {
		if a.CountAsPlay {
			plays = append(plays, a)
		}
	}
	if replaceSinglePlay && len(plays) == 1 {
		slog.Debug("applyImportWatches: Replacing single existing play.",
			"activity_id", plays[0].ID)
		res := s.db.Where("user_id = ?", userId).Delete(&entity.Activity{}, plays[0].ID)
		if res.Error != nil {
			return errors.New("failed removing replaced play activity")
		}
		plays = []entity.Activity{}
	}
	for _, watch := range watches {
		if findPlayOnDay(plays, watch.Date) != nil {
			continue
		}
		customDate := watch.Date
		added, err := s.activityProvider.AddActivity(
			userId,
			domain.ActivityAddProps{
				WatchedID:  watchedId,
				Type:       entity.IMPORTED_ADDED_WATCHED,
				CustomDate: &customDate,
			},
			true,
		)
		if err != nil {
			slog.Error("applyImportWatches: Failed to add play activity.",
				"date", watch.Date, "error", err)
			continue
		}
		plays = append(plays, added)
	}
	return nil
}

// Find an existing play on the same calendar day as `date`.
func findPlayOnDay(plays []entity.Activity, date time.Time) *entity.Activity {
	for i := range plays {
		playDate := plays[i].CreatedAt
		if plays[i].CustomDate != nil && !plays[i].CustomDate.IsZero() {
			playDate = *plays[i].CustomDate
		}
		y1, m1, d1 := playDate.UTC().Date()
		y2, m2, d2 := date.UTC().Date()
		if y1 == y2 && m1 == m2 && d1 == d2 {
			return &plays[i]
		}
	}
	return nil
}
