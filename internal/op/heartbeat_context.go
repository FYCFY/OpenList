package op

import (
	"context"
	"fmt"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
)

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
