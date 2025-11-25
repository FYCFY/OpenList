package davsession

import (
	"errors"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/utils/random"
	"gorm.io/gorm"
)

const sessionIdleTTL = 1 * time.Minute

func cleanStale(now time.Time, d *gorm.DB) {
	// Remove sessions that have been inactive for a while to keep the list “live”.
	_ = d.Where("last_seen < ?", now.Add(-sessionIdleTTL)).Delete(&model.WebdavSession{}).Error
}

// EnsureSession registers or refreshes a WebDAV session for the given user/ip/ua.
// Returns allowed=false when user-level limit is hit or the session was force-closed.
func EnsureSession(user *model.User, ip, ua string) (*model.WebdavSession, bool, error) {
	now := time.Now()
	d := db.GetDb()
	cleanStale(now, d)
	var session model.WebdavSession
	err := d.Where("user_id = ? AND ip = ?", user.ID, ip).First(&session).Error
	if err == nil {
		if session.ForceClose {
			// Remove stale forced sessions so the user can re-login after being kicked.
			_ = d.Delete(&session).Error
			return nil, false, nil
		}
		_ = d.Model(&session).Updates(map[string]interface{}{
			"last_seen":   now,
			"updated_at":  now,
			"user_agent":  ua,
			"force_close": false,
		}).Error
		return &session, true, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	limit := user.WebdavMaxSessions
	if limit > 0 {
		var ipCount int64
		if err := d.Model(&model.WebdavSession{}).
			Where("user_id = ? AND force_close = ?", user.ID, false).
			Distinct("ip").
			Count(&ipCount).Error; err != nil {
			return nil, false, err
		}
		if ipCount >= int64(limit) {
			return nil, false, nil
		}
	}

	session = model.WebdavSession{
		SessionID: random.String(24),
		UserID:    user.ID,
		Username:  user.Username,
		IP:        ip,
		UserAgent: ua,
		LastSeen:  now,
	}
	if err := d.Create(&session).Error; err != nil {
		return nil, false, err
	}
	return &session, true, nil
}

// ListSessions returns paginated sessions with optional username filter.
func ListSessions(username string, page, perPage int) ([]model.WebdavSession, int64, error) {
	d := db.GetDb()
	now := time.Now()
	cleanStale(now, d)
	query := d.Model(&model.WebdavSession{})
	if username != "" {
		query = query.Where("username = ?", username)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	var sessions []model.WebdavSession
	if err := query.Order("last_seen DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

// ForceCloseSession marks a session as force-closed and drops it from store.
func ForceCloseSession(id uint) error {
	d := db.GetDb()
	return d.Delete(&model.WebdavSession{}, id).Error
}
