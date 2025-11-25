package op

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"gorm.io/gorm"
)

// UpsertDevice updates existing device by android_id or creates a new record.
func UpsertDevice(user *model.User, payload *model.Device) (*model.Device, error) {
	now := time.Now()
	d := db.GetDb()
	var device model.Device
	err := d.Where("android_id = ?", payload.AndroidID).First(&device).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	payload.UserID = &user.ID
	payload.Username = user.Username
	payload.LastSeen = &now
	payload.LastIP = payload.LastIP
	payload.LastUserAgent = payload.LastUserAgent

	if err == gorm.ErrRecordNotFound {
		payload.FirstSeen = now
		payload.CreatedAt = now
		payload.UpdatedAt = now
		if err := d.Create(payload).Error; err != nil {
			return nil, err
		}
		return payload, nil
	}

	updates := map[string]interface{}{
		"device_codename": payload.DeviceCodename,
		"android_version": payload.AndroidVersion,
		"system_version":  payload.SystemVersion,
		"device_model":    payload.DeviceModel,
		"device_brand":    payload.DeviceBrand,
		"device_serial":   payload.DeviceSerial,
		"rom_version":     payload.ROMVersion,
		"build_date":      payload.BuildDate,
		"security_patch":  payload.SecurityPatch,
		"kernel_version":  payload.KernelVersion,
		"total_ram":       payload.TotalRAM,
		"storage_info":    payload.StorageInfo,
		"last_ip":         payload.LastIP,
		"last_user_agent": payload.LastUserAgent,
		"username":        payload.Username,
		"user_id":         payload.UserID,
		"last_seen":       payload.LastSeen,
		"updated_at":      now,
	}
	if payload.Remark != "" {
		updates["remark"] = payload.Remark
	}
	if err := d.Model(&device).Updates(updates).Error; err != nil {
		return nil, err
	}
	if err := d.Where("id = ?", device.ID).First(&device).Error; err != nil {
		return nil, err
	}
	return &device, nil
}

type DeviceFilter struct {
	Username string
	Search   string
	Start    *time.Time
	End      *time.Time
	Page     int
	PerPage  int
}

func ListDevices(filter DeviceFilter) ([]model.Device, int64, error) {
	var devices []model.Device
	query := db.GetDb().Model(&model.Device{})
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where(
			"android_id LIKE ? OR device_codename LIKE ? OR device_model LIKE ? OR device_brand LIKE ?",
			like, like, like, like,
		)
	}
	if filter.Start != nil {
		query = query.Where("last_seen >= ?", *filter.Start)
	}
	if filter.End != nil {
		query = query.Where("last_seen <= ?", *filter.End)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if err := query.Order("last_seen DESC NULLS LAST, first_seen DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&devices).Error; err != nil {
		return nil, 0, err
	}
	return devices, total, nil
}

func DeleteDevices(ids []uint) error {
	return db.GetDb().Where("id IN ?", ids).Delete(&model.Device{}).Error
}

func UpdateDeviceRemark(id uint, remark string) error {
	return db.GetDb().Model(&model.Device{}).Where("id = ?", id).Update("remark", remark).Error
}
