package op

import (
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/cron"
	log "github.com/sirupsen/logrus"
)

const (
	settingCleanupEnable      = "monitor_cleanup_enable"
	settingCleanupMaxAgeDays  = "monitor_cleanup_max_age_days"
	settingCleanupInactiveDay = "monitor_cleanup_inactive_days"
)

type DeviceCleanupConfig struct {
	Enable       bool `json:"enable"`
	MaxAgeDays   int  `json:"max_age_days"`
	InactiveDays int  `json:"inactive_days"`
}

func GetDeviceCleanupConfig() DeviceCleanupConfig {
	return DeviceCleanupConfig{
		Enable:       getSettingBool(settingCleanupEnable),
		MaxAgeDays:   getSettingInt(settingCleanupMaxAgeDays),
		InactiveDays: getSettingInt(settingCleanupInactiveDay),
	}
}

func SaveDeviceCleanupConfig(cfg DeviceCleanupConfig) error {
	items := []model.SettingItem{
		{Key: settingCleanupEnable, Value: boolToStr(cfg.Enable), Type: conf.TypeBool, Group: model.PRIVATE},
		{Key: settingCleanupMaxAgeDays, Value: intToStr(cfg.MaxAgeDays), Type: conf.TypeNumber, Group: model.PRIVATE},
		{Key: settingCleanupInactiveDay, Value: intToStr(cfg.InactiveDays), Type: conf.TypeNumber, Group: model.PRIVATE},
	}
	return SaveSettingItems(items)
}

// RunDeviceCleanup deletes devices that match configured rules.
func RunDeviceCleanup() error {
	cfg := GetDeviceCleanupConfig()
	if !cfg.Enable {
		return nil
	}
	now := time.Now()
	idMap := make(map[uint]struct{})
	if cfg.MaxAgeDays > 0 {
		cutoff := now.Add(-time.Duration(cfg.MaxAgeDays) * 24 * time.Hour)
		var ids []uint
		_ = db.GetDb().Model(&model.Device{}).Where("first_seen < ?", cutoff).Pluck("id", &ids).Error
		for _, id := range ids {
			idMap[id] = struct{}{}
		}
	}
	if cfg.InactiveDays > 0 {
		cutoff := now.Add(-time.Duration(cfg.InactiveDays) * 24 * time.Hour)
		var ids []uint
		_ = db.GetDb().Model(&model.Device{}).
			Where("(last_seen IS NULL AND first_seen < ?) OR (last_seen < ?)", cutoff, cutoff).
			Pluck("id", &ids).Error
		for _, id := range ids {
			idMap[id] = struct{}{}
		}
	}
	if len(idMap) == 0 {
		return nil
	}
	var delIDs []uint
	for id := range idMap {
		delIDs = append(delIDs, id)
	}
	if err := DeleteDevices(delIDs); err != nil {
		return err
	}
	log.Infof("自动清理设备完成，删除数量: %d", len(delIDs))
	return nil
}

var cleanupCron *cron.Cron

// StartDeviceCleanupScheduler launches periodic cleanup (hourly).
func StartDeviceCleanupScheduler() {
	if cleanupCron != nil {
		return
	}
	cleanupCron = cron.NewCron(time.Hour)
	cleanupCron.Do(func() {
		if err := RunDeviceCleanup(); err != nil {
			log.Warnf("自动清理设备失败: %v", err)
		}
	})
	// 立即执行一次，避免等待首个周期
	go func() {
		if err := RunDeviceCleanup(); err != nil {
			log.Warnf("初始自动清理设备失败: %v", err)
		}
	}()
}

func StopDeviceCleanupScheduler() {
	if cleanupCron != nil {
		cleanupCron.Stop()
	}
}

func getSettingInt(key string) int {
	s := getSettingStr(key)
	i, _ := strconv.Atoi(s)
	return i
}

func intToStr(i int) string {
	return strconv.Itoa(i)
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
