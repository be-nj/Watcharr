package source

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

// Personal cinema stats: visits per year (with the snack rating trend)
// and a ranking of the visited cinemas. A visit is a play that has a
// cinema source attached to it.
func (s *Service) GetCinemaStats(userId uint) (domain.CinemaStatsResponse, error) {
	stats := domain.CinemaStatsResponse{
		Years:   []domain.CinemaStatsYear{},
		Ranking: []domain.CinemaStatsRankingEntry{},
	}
	res := s.db.
		Table("activities a").
		Select(`strftime('%Y', COALESCE(a.custom_date, a.created_at)) AS year,
			COUNT(*) AS visits,
			AVG(ad.rating_snacks) AS rating_snacks_avg`).
		Joins("JOIN activity_details ad ON ad.activity_id = a.id").
		Joins("JOIN watch_sources ws ON ws.id = ad.watch_source_id AND ws.type = ? AND ws.deleted_at IS NULL",
			entity.SOURCE_CINEMA).
		Where("a.user_id = ? AND a.count_as_play = 1 AND a.deleted_at IS NULL", userId).
		Group("year").
		Order("year").
		Scan(&stats.Years)
	if res.Error != nil {
		slog.Error("getCinemaStats: Failed getting visits per year",
			"error", res.Error.Error())
		return stats, errors.New("failed getting visits per year")
	}
	res = s.db.
		Table("activities a").
		Select(`ws.id, ws.name, cd.city,
			COUNT(*) AS visits,
			AVG(ad.rating_overall) AS rating_overall_avg`).
		Joins("JOIN activity_details ad ON ad.activity_id = a.id").
		Joins("JOIN watch_sources ws ON ws.id = ad.watch_source_id AND ws.type = ? AND ws.deleted_at IS NULL",
			entity.SOURCE_CINEMA).
		Joins("LEFT JOIN cinema_details cd ON cd.watch_source_id = ws.id").
		Where("a.user_id = ? AND a.count_as_play = 1 AND a.deleted_at IS NULL", userId).
		Group("ws.id").
		Order("visits DESC, ws.name").
		Scan(&stats.Ranking)
	if res.Error != nil {
		slog.Error("getCinemaStats: Failed getting cinema ranking",
			"error", res.Error.Error())
		return stats, errors.New("failed getting cinema ranking")
	}
	stats.DistinctCinemas = len(stats.Ranking)
	for _, r := range stats.Ranking {
		stats.TotalVisits += r.Visits
	}
	return stats, nil
}
