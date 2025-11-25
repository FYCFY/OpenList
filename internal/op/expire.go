package op

import (
	"context"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	log "github.com/sirupsen/logrus"
)

func isUserExpired(u *model.User, now time.Time) bool {
	return u.ExpiresAt != nil && now.After(*u.ExpiresAt) && !u.IsAdmin() && !u.IsGuest()
}

// CleanupExpiredUsers deletes expired users (non admin/guest).
func CleanupExpiredUsers() (int64, error) {
	now := time.Now().UTC()
	res := db.GetDb().Where("expires_at IS NOT NULL AND expires_at < ? AND role NOT IN (?,?)", now, model.ADMIN, model.GUEST).Delete(&model.User{})
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected > 0 {
		log.Infof("removed %d expired users", res.RowsAffected)
	}
	return res.RowsAffected, nil
}

// PeriodicCleanExpiredUsers runs cleanup periodically until context canceled.
func PeriodicCleanExpiredUsers(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, _ = CleanupExpiredUsers()
		case <-ctx.Done():
			return
		}
	}
}
