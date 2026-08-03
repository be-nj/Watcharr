package export

import (
	"bytes"
	"encoding/csv"
	"errors"
	"log/slog"
	"strconv"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db}
}

type letterboxdRow struct {
	Title       string
	ReleaseDate string
	TmdbID      int
	Rating      float64
	WatchedDate string
}

// Build a csv in letterboxds import format (movies only, one row per
// play so rewatches survive; movies without a play date get one row
// without WatchedDate). Includes tmdbID so letterboxd matches exactly.
// https://letterboxd.com/about/importing-data/
func (s *Service) LetterboxdCsv(userId uint) ([]byte, error) {
	// One row per play (count_as_play activity)...
	plays := []letterboxdRow{}
	res := s.db.
		Table("watcheds w").
		Select(`c.title, c.release_date, c.tmdb_id, w.rating,
			strftime('%Y-%m-%d', COALESCE(a.custom_date, a.created_at)) AS watched_date`).
		Joins("JOIN contents c ON c.id = w.content_id AND c.type = 'movie'").
		Joins("JOIN activities a ON a.watched_id = w.id AND a.count_as_play = 1 AND a.deleted_at IS NULL").
		Where("w.user_id = ? AND w.deleted_at IS NULL", userId).
		Order("watched_date").
		Scan(&plays)
	if res.Error != nil {
		slog.Error("letterboxdCsv: Failed getting plays", "error", res.Error.Error())
		return nil, errors.New("failed getting plays")
	}
	// ...plus movies without any play.
	unplayed := []letterboxdRow{}
	res = s.db.
		Table("watcheds w").
		Select("c.title, c.release_date, c.tmdb_id, w.rating").
		Joins("JOIN contents c ON c.id = w.content_id AND c.type = 'movie'").
		Where(`w.user_id = ? AND w.deleted_at IS NULL AND w.status IN ('FINISHED', 'DROPPED', 'HOLD')
			AND NOT EXISTS (SELECT 1 FROM activities a WHERE a.watched_id = w.id
				AND a.count_as_play = 1 AND a.deleted_at IS NULL)`, userId).
		Scan(&unplayed)
	if res.Error != nil {
		slog.Error("letterboxdCsv: Failed getting unplayed", "error", res.Error.Error())
		return nil, errors.New("failed getting watched movies")
	}

	buf := &bytes.Buffer{}
	wr := csv.NewWriter(buf)
	_ = wr.Write([]string{"Title", "Year", "tmdbID", "Rating10", "WatchedDate"})
	writeRow := func(r letterboxdRow) {
		year := ""
		if len(r.ReleaseDate) >= 4 {
			year = r.ReleaseDate[:4]
		}
		rating := ""
		if r.Rating > 0 {
			rating = strconv.FormatFloat(r.Rating, 'f', -1, 64)
		}
		_ = wr.Write([]string{r.Title, year, strconv.Itoa(r.TmdbID), rating, r.WatchedDate})
	}
	for _, r := range plays {
		writeRow(r)
	}
	for _, r := range unplayed {
		writeRow(r)
	}
	wr.Flush()
	if err := wr.Error(); err != nil {
		return nil, errors.New("failed writing csv")
	}
	// Letterboxd caps import files at ~1900 rows; log so oversized
	// exports don't fail silently over there.
	if total := len(plays) + len(unplayed); total > 1800 {
		slog.Warn("letterboxdCsv: Export exceeds letterboxds import row limit, "+
			"the file may need manual splitting.", "rows", total)
	}
	return buf.Bytes(), nil
}
