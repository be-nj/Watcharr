package activity

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

// Log one of our plays for friends that watched together with us.
// Friends = mutual follows only; each of them gets their own play
// activity (FRIEND_LOGGED_WATCH, marked with our username) plus a copy
// of the watch details (source/screen/languages - never ratings or
// notes, those stay personal). The friend can simply delete the
// activity if they don't want it.
func (s *Service) LogCompanions(
	actorId uint,
	activityId uint,
	userIds []uint,
) ([]domain.CompanionLogResult, error) {
	if len(userIds) == 0 {
		return nil, errors.New("no users given")
	}
	// The play we are copying must be ours.
	activity := entity.Activity{}
	res := s.db.
		Where("id = ? AND user_id = ?", activityId, actorId).
		Preload("Details").
		Take(&activity)
	if res.Error != nil {
		return nil, errors.New("activity does not exist")
	}
	watched := entity.Watched{}
	res = s.db.Where("id = ?", activity.WatchedID).Take(&watched)
	if res.Error != nil {
		return nil, errors.New("watched entry does not exist")
	}
	if watched.ContentID == nil {
		return nil, errors.New("only movie/show watches can be logged for friends")
	}
	actor := entity.User{}
	if res := s.db.Where("id = ?", actorId).Take(&actor); res.Error != nil {
		return nil, errors.New("failed getting own user")
	}
	activityData, err := json.Marshal(map[string]string{
		"loggedBy": actor.Username,
	})
	if err != nil {
		return nil, errors.New("failed marshalling activity data")
	}
	playDate := activity.CreatedAt
	if activity.CustomDate != nil {
		playDate = *activity.CustomDate
	}

	results := []domain.CompanionLogResult{}
	for _, targetId := range userIds {
		result := domain.CompanionLogResult{UserID: targetId}
		target := entity.User{}
		if res := s.db.Where("id = ?", targetId).Take(&target); res.Error != nil {
			result.Error = "user does not exist"
			results = append(results, result)
			continue
		}
		result.Username = target.Username
		if targetId == actorId {
			result.Error = "cannot log for yourself"
			results = append(results, result)
			continue
		}
		if !s.isMutualFollow(actorId, targetId) {
			result.Error = "not a mutual follow"
			results = append(results, result)
			continue
		}
		if err := s.logCompanionPlay(
			targetId, watched, playDate, string(activityData), activity.Details,
		); err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}
		result.Ok = true
		results = append(results, result)
	}
	return results, nil
}

// Both directions must exist for users to count as friends.
func (s *Service) isMutualFollow(a uint, b uint) bool {
	var count int64
	res := s.db.Model(&entity.Follow{}).
		Where("(user_id = ? AND followed_user_id = ?) OR (user_id = ? AND followed_user_id = ?)",
			a, b, b, a).
		Count(&count)
	if res.Error != nil {
		slog.Error("isMutualFollow: Failed counting follows", "error", res.Error.Error())
		return false
	}
	return count == 2
}

func (s *Service) logCompanionPlay(
	targetId uint,
	watched entity.Watched,
	playDate time.Time,
	activityData string,
	details *entity.ActivityDetails,
) error {
	// Get (or revive/create) the targets watched entry for the same
	// content. Created directly - the content row already exists (the
	// actor watched it), so no provider round trip is needed.
	targetWatched := entity.Watched{}
	res := s.db.Unscoped().
		Where("user_id = ? AND content_id = ?", targetId, *watched.ContentID).
		Take(&targetWatched)
	if res.Error == nil {
		if targetWatched.DeletedAt.Valid {
			if err := s.db.Unscoped().Model(&targetWatched).
				Update("deleted_at", nil).Error; err != nil {
				slog.Error("logCompanionPlay: Failed reviving watched entry",
					"error", err.Error())
				return errors.New("failed reviving watched entry")
			}
		}
	} else {
		targetWatched = entity.Watched{
			UserID:    targetId,
			ContentID: watched.ContentID,
			Status:    entity.FINISHED,
		}
		if err := s.db.Create(&targetWatched).Error; err != nil {
			slog.Error("logCompanionPlay: Failed creating watched entry",
				"error", err.Error())
			return errors.New("failed creating watched entry")
		}
	}
	act, err := s.AddActivity(targetId, domain.ActivityAddProps{
		WatchedID:  targetWatched.ID,
		Type:       entity.FRIEND_LOGGED_WATCH,
		Data:       activityData,
		CustomDate: &playDate,
	}, true)
	if err != nil {
		return errors.New("failed adding activity")
	}
	if details != nil {
		targetDetails := entity.ActivityDetails{
			ActivityID:     act.ID,
			WatchSourceID:  details.WatchSourceID,
			CinemaScreenID: details.CinemaScreenID,
			AudioLang:      details.AudioLang,
			SubtitleLang:   details.SubtitleLang,
		}
		if err := s.db.Create(&targetDetails).Error; err != nil {
			// The play itself made it - don't fail the whole log.
			slog.Error("logCompanionPlay: Failed copying watch details",
				"error", err.Error())
		}
	}
	return nil
}
