package model

import "time"

// WebdavBlock represents a blocked IP for WebDAV access.
type WebdavBlock struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	IP        string     `json:"ip" gorm:"uniqueIndex;size:64"`
	Remark    string     `json:"remark" gorm:"size:512"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
