package op

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
)

const (
	settingHeartbeatEnable = "monitor_heartbeat_enable"
	settingHeartbeatUser   = "monitor_heartbeat_user"
	settingHeartbeatScript = "monitor_heartbeat_script"
)

type HeartbeatConfig struct {
	Enable   bool   `json:"enable"`
	Username string `json:"username"`
	Script   string `json:"script"`
}

func GetHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Enable:   getSettingBool(settingHeartbeatEnable),
		Username: getSettingStr(settingHeartbeatUser),
		Script:   getSettingStr(settingHeartbeatScript),
	}
}

func SaveHeartbeatConfig(cfg HeartbeatConfig) error {
	items := []model.SettingItem{
		{Key: settingHeartbeatEnable, Value: fmt.Sprintf("%v", cfg.Enable), Type: conf.TypeBool, Group: model.PRIVATE},
		{Key: settingHeartbeatUser, Value: cfg.Username, Type: conf.TypeString, Group: model.PRIVATE},
		{Key: settingHeartbeatScript, Value: cfg.Script, Type: conf.TypeText, Group: model.PRIVATE},
	}
	return SaveSettingItems(items)
}

// uploadScriptWithConfig writes bl.sh to mounted storage using the configured user.
func uploadScriptWithConfig(ctx context.Context, device *model.Device, content string) error {
	if device == nil || device.AndroidID == "" {
		return fmt.Errorf("invalid device")
	}
	username := getSettingStr(settingHeartbeatUser)
	if username == "" {
		return fmt.Errorf("未配置心跳脚本用户")
	}
	hbUser, err := GetUserByName(username)
	if err != nil {
		return fmt.Errorf("心跳脚本用户不存在: %w", err)
	}
	dirPath := fmt.Sprintf("/sh/%s", device.AndroidID)
	ctx = context.WithValue(ctx, conf.UserKey, hbUser)
	if err := fs.MakeDir(ctx, dirPath); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	file := &model.FileStream{
		Obj: &model.Object{
			Name:     "bl.sh",
			Size:     int64(len(content)),
			Modified: time.Now(),
		},
		Reader:   strings.NewReader(content),
		Mimetype: "text/plain",
	}
	if err := fs.PutDirectly(ctx, dirPath, file, true); err != nil {
		return fmt.Errorf("写入脚本失败: %w", err)
	}
	return nil
}

func UploadDeviceScript(ctx context.Context, device *model.Device, content string) error {
	return uploadScriptWithConfig(ctx, device, content)
}

func ApplyDefaultHeartbeat(ctx context.Context, device *model.Device) error {
	if !getSettingBool(settingHeartbeatEnable) {
		return fmt.Errorf("未开启默认心跳脚本")
	}
	script := getSettingStr(settingHeartbeatScript)
	if script == "" {
		return fmt.Errorf("未配置心跳脚本内容")
	}
	return uploadScriptWithConfig(ctx, device, script)
}

func DeleteDeviceScript(ctx context.Context, device *model.Device) error {
	if device == nil || device.AndroidID == "" {
		return fmt.Errorf("invalid device")
	}
	username := getSettingStr(settingHeartbeatUser)
	if username == "" {
		return fmt.Errorf("未配置心跳脚本用户")
	}
	hbUser, err := GetUserByName(username)
	if err != nil {
		return fmt.Errorf("心跳脚本用户不存在: %w", err)
	}
	ctx = context.WithValue(ctx, conf.UserKey, hbUser)
	basePath := fmt.Sprintf("/sh/%s", device.AndroidID)
	return fs.Remove(ctx, basePath)
}

func getSettingStr(key string) string {
	if v, _ := GetSettingItemByKey(key); v != nil {
		return v.Value
	}
	return ""
}

func getSettingBool(key string) bool {
	s := getSettingStr(key)
	b, _ := strconv.ParseBool(s)
	return b
}
