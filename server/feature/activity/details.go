package activity

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

// Set/replace the details (watch source, language, tags, note) of one
// of the users activities. The details row is created lazily on first use.
func (s *Service) UpdateActivityDetails(
	userId uint,
	id uint,
	dr domain.ActivityDetailsUpdateRequest,
) (entity.ActivityDetails, error) {
	if id == 0 {
		return entity.ActivityDetails{},
			errors.New("id must be set to update activity details")
	}
	// Verify the user owns the activity.
	activity := new(entity.Activity)
	res := s.db.
		Model(&entity.Activity{}).
		Where("user_id = ? AND id = ?", userId, id).
		Find(&activity)
	if res.Error != nil {
		slog.Error("UpdateActivityDetails: Failed getting activity from database",
			"error", res.Error.Error())
		return entity.ActivityDetails{}, errors.New("failed getting activity")
	}
	if activity.ID == 0 {
		return entity.ActivityDetails{}, errors.New("activity does not exist")
	}
	// Verify the referenced source (and screen) belong to the user.
	if dr.WatchSourceID != nil {
		source := new(entity.WatchSource)
		res = s.db.
			Model(&entity.WatchSource{}).
			Where("id = ? AND user_id = ?", *dr.WatchSourceID, userId).
			Preload("Cinema").
			Preload("Cinema.Screens").
			Find(&source)
		if res.Error != nil {
			slog.Error("UpdateActivityDetails: Failed getting source from database",
				"error", res.Error.Error())
			return entity.ActivityDetails{}, errors.New("failed getting source")
		}
		if source.ID == 0 {
			return entity.ActivityDetails{}, errors.New("source does not exist")
		}
		if dr.CinemaScreenID != nil {
			if source.Cinema == nil {
				return entity.ActivityDetails{}, errors.New("source is not a cinema")
			}
			screenFound := false
			for _, screen := range source.Cinema.Screens {
				if screen.ID == *dr.CinemaScreenID {
					screenFound = true
					break
				}
			}
			if !screenFound {
				return entity.ActivityDetails{}, errors.New("screen does not exist")
			}
		}
	} else if dr.CinemaScreenID != nil {
		return entity.ActivityDetails{}, errors.New("screen cannot be set without a source")
	}
	// Get or create the details row.
	details := new(entity.ActivityDetails)
	res = s.db.Where("activity_id = ?", activity.ID).Find(&details)
	if res.Error != nil {
		slog.Error("UpdateActivityDetails: Failed getting details from database",
			"error", res.Error.Error())
		return entity.ActivityDetails{}, errors.New("failed getting activity details")
	}
	details.ActivityID = activity.ID
	details.WatchSourceID = dr.WatchSourceID
	details.CinemaScreenID = dr.CinemaScreenID
	details.AudioLang = dr.AudioLang
	details.SubtitleLang = dr.SubtitleLang
	details.Note = dr.Note
	res = s.db.Save(&details)
	if res.Error != nil {
		slog.Error("UpdateActivityDetails: Error saving details to database",
			"error", res.Error.Error())
		return entity.ActivityDetails{}, errors.New("failed saving activity details to database")
	}
	slog.Debug("UpdateActivityDetails: Updated details", "details", details)
	return *details, nil
}
