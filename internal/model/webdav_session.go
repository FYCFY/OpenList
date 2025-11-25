package model

import "time"

// WebdavSession records an authenticated WebDAV session for a user.
// Persisted so limits survive restarts and administrators can manage active sessions.
type WebdavSession struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"session_id" gorm:"uniqueIndex"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Username   string    `json:"username" gorm:"index"`
	IP         string    `json:"ip" gorm:"index"`
	UserAgent  string    `json:"user_agent" gorm:"size:512"`
	ForceClose bool      `json:"force_close" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastSeen   time.Time `json:"last_seen" gorm:"index"`
}
