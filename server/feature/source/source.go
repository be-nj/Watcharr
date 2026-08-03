package source

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db}
}

// Sources are shared by the whole instance (ADR 0003): everyone can
// list, create and edit them. Deleting is limited (see DeleteSource).

func (s *Service) GetSources(userId uint) ([]entity.WatchSource, error) {
	sources := new([]entity.WatchSource)
	res := s.db.
		Model(&entity.WatchSource{}).
		Preload("Cinema").
		Preload("Cinema.Screens").
		Find(&sources)
	if res.Error != nil {
		slog.Error("getSources: Failed getting sources from database", "error", res.Error.Error())
		return []entity.WatchSource{}, errors.New("failed getting sources")
	}
	for i := range *sources {
		s.fillRatingAggregate(&(*sources)[i])
	}
	return *sources, nil
}

func (s *Service) GetSource(userId uint, sourceId uint) (entity.WatchSource, error) {
	source := new(entity.WatchSource)
	res := s.db.
		Model(&entity.WatchSource{}).
		Where("id = ?", sourceId).
		Preload("Cinema").
		Preload("Cinema.Screens").
		Find(&source)
	if res.Error != nil {
		slog.Error("getSource: Failed getting source from database",
			"source_id", sourceId, "error", res.Error.Error())
		return entity.WatchSource{}, errors.New("failed getting source")
	}
	if source.ID == 0 {
		return entity.WatchSource{}, errors.New("source does not exist")
	}
	s.fillRatingAggregate(source)
	return *source, nil
}

// Average overall visit rating + rating count for a source (cinemas).
func (s *Service) fillRatingAggregate(source *entity.WatchSource) {
	if source.Type != entity.SOURCE_CINEMA {
		return
	}
	row := struct {
		Avg   *float64
		Count int64
	}{}
	res := s.db.
		Table("activity_details ad").
		Select("AVG(ad.rating_overall) AS avg, COUNT(ad.rating_overall) AS count").
		Joins("JOIN activities a ON a.id = ad.activity_id AND a.deleted_at IS NULL").
		Where("ad.watch_source_id = ? AND ad.rating_overall IS NOT NULL", source.ID).
		Scan(&row)
	if res.Error != nil {
		slog.Error("fillRatingAggregate: Failed", "source_id", source.ID,
			"error", res.Error.Error())
		return
	}
	source.RatingAverage = row.Avg
	source.RatingCount = row.Count
}

// Let a user create a source for the instance. Cinema sources carry
// their osm reference (empty for free form cinemas) and get an empty
// details row so screens/notes can be added afterwards.
func (s *Service) AddSource(userId uint, sr domain.WatchSourceAddRequest) (entity.WatchSource, error) {
	if sr.Name == "" {
		return entity.WatchSource{}, errors.New("source must have a name")
	}
	sourceType := entity.WatchSourceType(sr.Type)
	if !sourceType.IsValid() {
		return entity.WatchSource{}, errors.New("invalid source type")
	}
	source := entity.WatchSource{CreatedBy: userId, Name: sr.Name, Type: sourceType}
	if sourceType == entity.SOURCE_CINEMA {
		// A cinema with an osm reference must not be added twice.
		if sr.OsmID != 0 {
			existing := new(entity.CinemaDetails)
			res := s.db.
				Where("osm_type = ? AND osm_id = ?", sr.OsmType, sr.OsmID).
				Find(&existing)
			if res.Error == nil && existing.ID != 0 {
				return entity.WatchSource{}, errors.New("cinema already exists")
			}
		}
		source.Cinema = &entity.CinemaDetails{
			City:       sr.City,
			Address:    sr.Address,
			Lat:        sr.Lat,
			Lon:        sr.Lon,
			OsmType:    sr.OsmType,
			OsmID:      sr.OsmID,
			WikidataID: sr.WikidataID,
		}
	}
	res := s.db.Create(&source)
	if res.Error != nil {
		slog.Error("addSource: Error adding source to database", "error", res.Error.Error())
		return entity.WatchSource{}, errors.New("failed adding new source to database")
	}
	slog.Debug("addSource: Adding source", "added_source", source)
	return source, nil
}

// Update name/type of a source (shared, so any user may edit).
func (s *Service) UpdateSource(userId uint, sourceId uint, sr domain.WatchSourceAddRequest) error {
	if sr.Name == "" {
		return errors.New("source must have a name")
	}
	sourceType := entity.WatchSourceType(sr.Type)
	if !sourceType.IsValid() {
		return errors.New("invalid source type")
	}
	source := entity.WatchSource{Name: sr.Name, Type: sourceType}
	res := s.db.Where("id = ?", sourceId).Updates(&source)
	if res.Error != nil {
		slog.Error("updateSource: Error updating source in database", "error", res.Error.Error())
		return errors.New("failed updating source in database")
	}
	if res.RowsAffected == 0 {
		return errors.New("source does not exist")
	}
	if sourceType == entity.SOURCE_CINEMA {
		if err := s.ensureCinemaDetails(sourceId); err != nil {
			return err
		}
	}
	return nil
}

// Update the cinema details of a cinema source (shared).
func (s *Service) UpdateCinemaDetails(userId uint, sourceId uint, cr domain.CinemaDetailsUpdateRequest) error {
	source, err := s.GetSource(userId, sourceId)
	if err != nil {
		return err
	}
	if source.Cinema == nil {
		return errors.New("source is not a cinema")
	}
	res := s.db.
		Model(&entity.CinemaDetails{}).
		Where("id = ?", source.Cinema.ID).
		Updates(map[string]interface{}{
			"City":    cr.City,
			"Address": cr.Address,
			"Lat":     cr.Lat,
			"Lon":     cr.Lon,
			"Note":    cr.Note,
		})
	if res.Error != nil {
		slog.Error("updateCinemaDetails: Error updating details in database", "error", res.Error.Error())
		return errors.New("failed updating cinema details in database")
	}
	return nil
}

func (s *Service) AddScreen(userId uint, sourceId uint, sr domain.CinemaScreenAddRequest) (entity.CinemaScreen, error) {
	source, err := s.GetSource(userId, sourceId)
	if err != nil {
		return entity.CinemaScreen{}, err
	}
	if source.Cinema == nil {
		return entity.CinemaScreen{}, errors.New("source is not a cinema")
	}
	screen := entity.CinemaScreen{CinemaDetailsID: source.Cinema.ID, Name: sr.Name}
	res := s.db.Create(&screen)
	if res.Error != nil {
		slog.Error("addScreen: Error adding screen to database", "error", res.Error.Error())
		return entity.CinemaScreen{}, errors.New("failed adding new screen to database")
	}
	return screen, nil
}

func (s *Service) DeleteScreen(userId uint, sourceId uint, screenId uint) error {
	source, err := s.GetSource(userId, sourceId)
	if err != nil {
		return err
	}
	if source.Cinema == nil {
		return errors.New("source is not a cinema")
	}
	res := s.db.
		Where("id = ? AND cinema_details_id = ?", screenId, source.Cinema.ID).
		Delete(&entity.CinemaScreen{})
	if res.Error != nil {
		slog.Error("deleteScreen: Error deleting screen from database", "error", res.Error.Error())
		return errors.New("failed deleting screen from database")
	}
	if res.RowsAffected == 0 {
		return errors.New("screen does not exist")
	}
	return nil
}

// Delete a source. Allowed when no OTHER user has watches referencing
// it; otherwise only admins may delete (their watches lose the source).
func (s *Service) DeleteSource(userId uint, sourceId uint) error {
	if sourceId == 0 {
		return errors.New("no source id provided")
	}
	source, err := s.GetSource(userId, sourceId)
	if err != nil {
		return err
	}
	var othersUsing int64
	res := s.db.
		Table("activity_details ad").
		Joins("JOIN activities a ON a.id = ad.activity_id AND a.deleted_at IS NULL").
		Where("ad.watch_source_id = ? AND a.user_id != ?", sourceId, userId).
		Count(&othersUsing)
	if res.Error != nil {
		return errors.New("failed checking source usage")
	}
	if othersUsing > 0 && !s.userIsAdmin(userId) {
		return errors.New("source is used by other users, only admins can delete it")
	}
	// Detach the source from all watches referencing it.
	res = s.db.
		Model(&entity.ActivityDetails{}).
		Where("watch_source_id = ?", sourceId).
		Updates(map[string]interface{}{"watch_source_id": nil, "cinema_screen_id": nil})
	if res.Error != nil {
		return errors.New("failed detaching source from watches")
	}
	if source.Cinema != nil {
		s.db.Where("cinema_details_id = ?", source.Cinema.ID).Delete(&entity.CinemaScreen{})
		s.db.Delete(&entity.CinemaDetails{}, source.Cinema.ID)
	}
	res = s.db.Unscoped().Where("id = ?", sourceId).Delete(&entity.WatchSource{})
	if res.Error != nil {
		slog.Error("deleteSource: Error deleting source from database", "error", res.Error.Error())
		return errors.New("failed deleting source from database")
	}
	return nil
}

func (s *Service) userIsAdmin(userId uint) bool {
	var permissions int
	res := s.db.
		Table("users").
		Select("permissions").
		Where("id = ?", userId).
		Scan(&permissions)
	if res.Error != nil {
		return false
	}
	return permissions&entity.PERM_ADMIN != 0
}

func (s *Service) ensureCinemaDetails(sourceId uint) error {
	details := new(entity.CinemaDetails)
	res := s.db.Where("watch_source_id = ?", sourceId).Find(&details)
	if res.Error != nil {
		return errors.New("failed checking cinema details")
	}
	if details.ID == 0 {
		res = s.db.Create(&entity.CinemaDetails{WatchSourceID: sourceId})
		if res.Error != nil {
			return errors.New("failed creating cinema details")
		}
	}
	return nil
}
