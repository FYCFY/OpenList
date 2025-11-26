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

// BindHeartbeatUserToCtx attaches the configured heartbeat user to ctx for downstream fs operations.
func BindHeartbeatUserToCtx(ctx context.Context) context.Context {
	username := getSettingStr(settingHeartbeatUser)
	if username == "" {
		return ctx
	}
	hbUser, err := GetUserByName(username)
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, conf.UserKey, hbUser)
}

func ValidateHeartbeatUser() error {
	username := getSettingStr(settingHeartbeatUser)
	if username == "" {
		return fmt.Errorf("未配置心跳脚本用户")
	}
	_, err := GetUserByName(username)
	return err
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

// UploadScript writes script content to /sh/<device>/bl.sh using heartbeat user.
func UploadScript(ctx context.Context, device *model.Device, content string) error {
	if device == nil || device.AndroidID == "" {
		return fmt.Errorf("invalid device")
	}
	ctx = BindHeartbeatUserToCtx(ctx)
	dirPath := fmt.Sprintf("/sh/%s", device.AndroidID)
	if err := fs.MakeDir(ctx, dirPath); err != nil {
		return err
	}
	file := model.NewFileReader("bl.sh", int64(len(content)), time.Now(), "text/plain", strings.NewReader(content))
	return fs.PutDirectly(ctx, dirPath, file, true)
}
