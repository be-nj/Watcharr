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

func (s *Service) GetSources(userId uint) ([]entity.WatchSource, error) {
	sources := new([]entity.WatchSource)
	res := s.db.
		Model(&entity.WatchSource{}).
		Where("user_id = ?", userId).
		Preload("Cinema").
		Preload("Cinema.Screens").
		Find(&sources)
	if res.Error != nil {
		slog.Error("getSources: Failed getting sources from database", "error", res.Error.Error())
		return []entity.WatchSource{}, errors.New("failed getting sources")
	}
	return *sources, nil
}

func (s *Service) GetSource(userId uint, sourceId uint) (entity.WatchSource, error) {
	source := new(entity.WatchSource)
	res := s.db.
		Model(&entity.WatchSource{}).
		Where("id = ? AND user_id = ?", sourceId, userId).
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
	return *source, nil
}

// Let user create a watch source. Cinema sources get an empty
// details row so screens/ratings can be added afterwards.
func (s *Service) AddSource(userId uint, sr domain.WatchSourceAddRequest) (entity.WatchSource, error) {
	if sr.Name == "" {
		return entity.WatchSource{}, errors.New("source must have a name")
	}
	sourceType := entity.WatchSourceType(sr.Type)
	if !sourceType.IsValid() {
		return entity.WatchSource{}, errors.New("invalid source type")
	}
	source := entity.WatchSource{UserID: userId, Name: sr.Name, Type: sourceType}
	if sourceType == entity.SOURCE_CINEMA {
		source.Cinema = &entity.CinemaDetails{}
	}
	res := s.db.Create(&source)
	if res.Error != nil {
		slog.Error("addSource: Error adding source to database", "error", res.Error.Error())
		return entity.WatchSource{}, errors.New("failed adding new source to database")
	}
	slog.Debug("addSource: Adding source", "added_source", source)
	return source, nil
}

// Let user update name/type of one of their sources (replaces).
// Changing a cinema into another type keeps the cinema details row
// around (harmless), changing into a cinema creates one if missing.
func (s *Service) UpdateSource(userId uint, sourceId uint, sr domain.WatchSourceAddRequest) error {
	if sr.Name == "" {
		return errors.New("source must have a name")
	}
	sourceType := entity.WatchSourceType(sr.Type)
	if !sourceType.IsValid() {
		return errors.New("invalid source type")
	}
	source := entity.WatchSource{Name: sr.Name, Type: sourceType}
	res := s.db.Where("id = ? AND user_id = ?", sourceId, userId).Updates(&source)
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

// Let user update the cinema details of one of their cinema sources.
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
			"City":          cr.City,
			"Address":       cr.Address,
			"Lat":           cr.Lat,
			"Lon":           cr.Lon,
			"Note":          cr.Note,
			"RatingOverall": cr.RatingOverall,
			"RatingSnacks":  cr.RatingSnacks,
			"RatingTech":    cr.RatingTech,
			"RatingComfort": cr.RatingComfort,
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

// Let user delete their own source.
func (s *Service) DeleteSource(userId uint, sourceId uint) error {
	if sourceId == 0 {
		return errors.New("no source id provided")
	}
	source, err := s.GetSource(userId, sourceId)
	if err != nil {
		return err
	}
	if source.Cinema != nil {
		s.db.Where("cinema_details_id = ?", source.Cinema.ID).Delete(&entity.CinemaScreen{})
		s.db.Delete(&entity.CinemaDetails{}, source.Cinema.ID)
	}
	res := s.db.Unscoped().Where("id = ? AND user_id = ?", sourceId, userId).Delete(&entity.WatchSource{})
	if res.Error != nil {
		slog.Error("deleteSource: Error deleting source from database", "error", res.Error.Error())
		return errors.New("failed deleting source from database")
	}
	return nil
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
