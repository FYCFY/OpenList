package model

import "time"

// Device holds client-reported device information.
type Device struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	AndroidID      string     `json:"android_id" gorm:"size:128;uniqueIndex"`
	DeviceCodename string     `json:"device_codename" gorm:"size:128;index"`
	AndroidVersion string     `json:"android_version" gorm:"size:64"`
	SystemVersion  string     `json:"system_version" gorm:"size:64"`
	DeviceModel    string     `json:"device_model" gorm:"size:128"`
	DeviceBrand    string     `json:"device_brand" gorm:"size:128"`
	DeviceSerial   string     `json:"device_serial" gorm:"size:128"`
	ROMVersion     string     `json:"rom_version" gorm:"size:128"`
	BuildDate      string     `json:"build_date" gorm:"size:128"`
	SecurityPatch  string     `json:"security_patch" gorm:"size:64"`
	KernelVersion  string     `json:"kernel_version" gorm:"size:128"`
	TotalRAM       string     `json:"total_ram" gorm:"size:64"`
	StorageInfo    string     `json:"storage_info" gorm:"size:128"`
	LastIP         string     `json:"last_ip" gorm:"size:64;index"`
	LastUserAgent  string     `json:"last_user_agent" gorm:"size:512"`
	Username       string     `json:"username" gorm:"size:128;index"`
	UserID         *uint      `json:"user_id" gorm:"index"`
	Remark         string     `json:"remark" gorm:"size:255"`
	FirstSeen      time.Time  `json:"first_seen" gorm:"index"`
	LastSeen       *time.Time `json:"last_seen" gorm:"index"`
	CreatedAt      time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"index"`
}
