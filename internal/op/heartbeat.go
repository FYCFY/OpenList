package op

import (
	"fmt"
	"strconv"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/pkg/gowebdav"
)

const (
	settingHeartbeatEnable   = "monitor_heartbeat_enable"
	settingHeartbeatUser     = "monitor_heartbeat_user"
	settingHeartbeatPassword = "monitor_heartbeat_password"
	settingHeartbeatScript   = "monitor_heartbeat_script"
)

type HeartbeatConfig struct {
	Enable   bool   `json:"enable"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Script   string `json:"script"`
}

func GetHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Enable:   getSettingBool(settingHeartbeatEnable),
		Username: getSettingStr(settingHeartbeatUser),
		Password: "", // 不回传
		Script:   getSettingStr(settingHeartbeatScript),
	}
}

func SaveHeartbeatConfig(cfg HeartbeatConfig) error {
	items := []model.SettingItem{
		{Key: settingHeartbeatEnable, Value: fmt.Sprintf("%v", cfg.Enable), Type: conf.TypeBool, Group: model.PRIVATE},
		{Key: settingHeartbeatUser, Value: cfg.Username, Type: conf.TypeString, Group: model.PRIVATE},
		{Key: settingHeartbeatPassword, Value: cfg.Password, Type: conf.TypeString, Group: model.PRIVATE},
		{Key: settingHeartbeatScript, Value: cfg.Script, Type: conf.TypeText, Group: model.PRIVATE},
	}
	return SaveSettingItems(items)
}

func uploadScriptWithConfig(device *model.Device, content string) error {
	if device == nil || device.AndroidID == "" {
		return fmt.Errorf("invalid device")
	}
	username := getSettingStr(settingHeartbeatUser)
	password := getSettingStr(settingHeartbeatPassword)
	if username == "" || password == "" {
		return fmt.Errorf("未配置心跳脚本账号或密码")
	}
	client := gowebdav.NewClient("", username, password)
	basePath := fmt.Sprintf("sh/%s", device.AndroidID)
	if err := client.MkdirAll(basePath, 0755); err != nil {
		return err
	}
	remote := fmt.Sprintf("%s/bl.sh", basePath)
	return client.Write(remote, []byte(content), 0644)
}

func UploadDeviceScript(device *model.Device, content string) error {
	return uploadScriptWithConfig(device, content)
}

func ApplyDefaultHeartbeat(device *model.Device) error {
	if !getSettingBool(settingHeartbeatEnable) {
		return fmt.Errorf("未开启默认心跳脚本")
	}
	script := getSettingStr(settingHeartbeatScript)
	if script == "" {
		return fmt.Errorf("未配置心跳脚本内容")
	}
	return uploadScriptWithConfig(device, script)
}

func DeleteDeviceScript(device *model.Device) error {
	if device == nil || device.AndroidID == "" {
		return fmt.Errorf("invalid device")
	}
	username := getSettingStr(settingHeartbeatUser)
	password := getSettingStr(settingHeartbeatPassword)
	if username == "" || password == "" {
		return fmt.Errorf("未配置心跳脚本账号或密码")
	}
	client := gowebdav.NewClient("", username, password)
	basePath := fmt.Sprintf("sh/%s", device.AndroidID)
	return client.RemoveAll(basePath)
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
