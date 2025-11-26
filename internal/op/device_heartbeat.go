package op

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/model"
)

// UpsertDeviceHeartbeat updates last seen info and basic metadata.
func UpsertDeviceHeartbeat(user *model.User, payload *model.Device) (*model.Device, error) {
	now := time.Now()
	payload.LastSeen = &now
	payload.LastIP = payload.LastIP
	payload.LastUserAgent = payload.LastUserAgent
	payload.OnlineStatus = "online"
	return UpsertDevice(user, payload)
}
