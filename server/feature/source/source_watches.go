package source

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/domain"
)

// Get the requesting users watches (activities) that used a source,
// newest first, with their content, for the source detail page. Only
// the own watches: other peoples viewing history stays private, their
// ratings surface via GetSourceRatings (anonymised unless opted in).
func (s *Service) GetSourceWatches(userId uint, sourceId uint) ([]domain.WatchSourceWatch, error) {
	if _, err := s.GetSource(userId, sourceId); err != nil {
		return nil, err
	}
	watches := []domain.WatchSourceWatch{}
	res := s.db.
		Table("activity_details ad").
		Select(`a.id AS activity_id,
			a.watched_id,
			COALESCE(a.custom_date, a.created_at) AS date,
			COALESCE(cs.name, '') AS screen_name,
			c.title,
			c.type,
			c.tmdb_id,
			c.poster_path`).
		Joins("JOIN activities a ON a.id = ad.activity_id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN cinema_screens cs ON cs.id = ad.cinema_screen_id").
		Joins("JOIN watcheds w ON w.id = a.watched_id AND w.deleted_at IS NULL").
		Joins("JOIN contents c ON c.id = w.content_id").
		Where("ad.watch_source_id = ? AND a.user_id = ?", sourceId, userId).
		Order("date DESC").
		Scan(&watches)
	if res.Error != nil {
		slog.Error("getSourceWatches: Failed getting watches from database",
			"source_id", sourceId, "error", res.Error.Error())
		return nil, errors.New("failed getting source watches")
	}
	return watches, nil
}

// All visit ratings of a source (cinema), newest first. The username
// is only included when the rater opted in (per rating); everyone
// elses ratings show anonymised.
func (s *Service) GetSourceRatings(userId uint, sourceId uint) ([]domain.WatchSourceRating, error) {
	if _, err := s.GetSource(userId, sourceId); err != nil {
		return nil, err
	}
	ratings := []domain.WatchSourceRating{}
	res := s.db.
		Table("activity_details ad").
		Select(`COALESCE(a.custom_date, a.created_at) AS date,
			ad.rating_overall,
			ad.rating_snacks,
			ad.rating_tech,
			ad.rating_comfort,
			CASE WHEN ad.rating_show_name THEN u.username ELSE '' END AS username,
			a.user_id = `+"?"+` AS own`, userId).
		Joins("JOIN activities a ON a.id = ad.activity_id AND a.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = a.user_id").
		Where("ad.watch_source_id = ? AND ad.rating_overall IS NOT NULL", sourceId).
		Order("date DESC").
		Scan(&ratings)
	if res.Error != nil {
		slog.Error("getSourceRatings: Failed getting ratings from database",
			"source_id", sourceId, "error", res.Error.Error())
		return nil, errors.New("failed getting source ratings")
	}
	return ratings, nil
}
