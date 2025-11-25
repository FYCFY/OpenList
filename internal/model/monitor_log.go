package model

import "time"

// LoginLog records a client login event coming from the WebDAV credentials.
type LoginLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"user_id" gorm:"index"`
	Username  string    `json:"username" gorm:"size:128;index"`
	IP        string    `json:"ip" gorm:"size:64;index"`
	UserAgent string    `json:"user_agent" gorm:"size:512"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// UploadLog records a client upload event.
type UploadLog struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        *uint     `json:"user_id" gorm:"index"`
	Username      string    `json:"username" gorm:"size:128;index"`
	DeviceCode    string    `json:"device_code" gorm:"size:128;index"`
	SystemVersion string    `json:"system_version" gorm:"size:64"`
	FileName      string    `json:"file_name" gorm:"size:512"`
	FileSize      int64     `json:"file_size"`
	IP            string    `json:"ip" gorm:"size:64;index"`
	CreatedAt     time.Time `json:"created_at" gorm:"index"`
}

// SystemLog records generic client side events.
type SystemLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"user_id" gorm:"index"`
	Username  string    `json:"username" gorm:"size:128;index"`
	Type      string    `json:"type" gorm:"size:32;index"` // info|warning|error
	Message   string    `json:"message" gorm:"type:text"`
	Source    string    `json:"source" gorm:"size:128"`
	IP        string    `json:"ip" gorm:"size:64;index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
