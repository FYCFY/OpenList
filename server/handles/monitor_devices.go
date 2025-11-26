package handles

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/fs"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/stream"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type deviceUpsertReq struct {
	AndroidID      string `json:"android_id" binding:"required"`
	DeviceCodename string `json:"device_codename"`
	AndroidVersion string `json:"android_version"`
	SystemVersion  string `json:"system_version"`
	DeviceModel    string `json:"device_model"`
	DeviceBrand    string `json:"device_brand"`
	DeviceSerial   string `json:"device_serial"`
	ROMVersion     string `json:"rom_version"`
	BuildDate      string `json:"build_date"`
	SecurityPatch  string `json:"security_patch"`
	KernelVersion  string `json:"kernel_version"`
	TotalRAM       string `json:"total_ram"`
	StorageInfo    string `json:"storage_info"`
	Remark         string `json:"remark"`
}

func UpsertDevice(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req deviceUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	now := time.Now()
	device := &model.Device{
		AndroidID:      req.AndroidID,
		DeviceCodename: req.DeviceCodename,
		AndroidVersion: req.AndroidVersion,
		SystemVersion:  req.SystemVersion,
		DeviceModel:    req.DeviceModel,
		DeviceBrand:    req.DeviceBrand,
		DeviceSerial:   req.DeviceSerial,
		ROMVersion:     req.ROMVersion,
		BuildDate:      req.BuildDate,
		SecurityPatch:  req.SecurityPatch,
		KernelVersion:  req.KernelVersion,
		TotalRAM:       req.TotalRAM,
		StorageInfo:    req.StorageInfo,
		Remark:         req.Remark,
		LastIP:         c.ClientIP(),
		LastUserAgent:  c.GetHeader("User-Agent"),
		LastSeen:       &now,
	}

	saved, err := op.UpsertDevice(user, device)
	if err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"success": true,
		"data":    saved,
	})
}

func AdminListDevices(c *gin.Context) {
	username := c.Query("username")
	search := c.Query("search")
	start, end, err := parseRange(c)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	devices, total, err := op.ListDevices(op.DeviceFilter{
		Username: username,
		Search:   search,
		Start:    start,
		End:      end,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: devices,
		Total:   total,
	})
}

type heartbeatConfigReq struct {
	Enable   bool   `json:"enable"`
	Username string `json:"username"`
	Script   string `json:"script"`
}

func GetHeartbeatConfig(c *gin.Context) {
	common.SuccessResp(c, op.GetHeartbeatConfig())
}

func SaveHeartbeatConfig(c *gin.Context) {
	var req heartbeatConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if err := op.SaveHeartbeatConfig(op.HeartbeatConfig{
		Enable:   req.Enable,
		Username: req.Username,
		Script:   req.Script,
	}); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	common.SuccessResp(c)
}

type deviceScriptReq struct {
	ID      uint   `json:"id" binding:"required"`
	Content string `json:"content"`
}

func UploadDeviceScriptHandle(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req deviceScriptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	device, err := op.GetDeviceByID(req.ID)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		common.ErrorStrResp(c, "脚本内容不能为空", http.StatusBadRequest)
		return
	}
	if err := op.ValidateHeartbeatUser(); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	ctx := op.BindHeartbeatUserToCtx(c.Request.Context())
	dirPath := fmt.Sprintf("/sh/%s", device.AndroidID)
	if err := fs.MakeDir(ctx, dirPath); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	file := &stream.FileStream{
		Obj: &model.Object{
			Name:     "bl.sh",
			Size:     int64(len(req.Content)),
			Modified: time.Now(),
		},
		Reader:   strings.NewReader(req.Content),
		Mimetype: "text/plain",
	}
	if err := fs.PutDirectly(ctx, dirPath, file, true); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	_ = op.AddSystemLog(user, model.SystemLog{
		Type:    "info",
		Message: fmt.Sprintf("上传设备脚本: %s", device.AndroidID),
		Source:  "device_script",
		IP:      c.ClientIP(),
	})
	common.SuccessResp(c)
}

func ApplyHeartbeatHandle(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req deviceScriptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	device, err := op.GetDeviceByID(req.ID)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	cfg := op.GetHeartbeatConfig()
	if !cfg.Enable {
		common.ErrorStrResp(c, "未开启默认心跳脚本", http.StatusBadRequest)
		return
	}
	if cfg.Script == "" {
		common.ErrorStrResp(c, "未配置心跳脚本内容", http.StatusBadRequest)
		return
	}
	if err := op.ValidateHeartbeatUser(); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	ctx := op.BindHeartbeatUserToCtx(c.Request.Context())
	dirPath := fmt.Sprintf("/sh/%s", device.AndroidID)
	if err := fs.MakeDir(ctx, dirPath); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	file := &stream.FileStream{
		Obj: &model.Object{
			Name:     "bl.sh",
			Size:     int64(len(cfg.Script)),
			Modified: time.Now(),
		},
		Reader:   strings.NewReader(cfg.Script),
		Mimetype: "text/plain",
	}
	if err := fs.PutDirectly(ctx, dirPath, file, true); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	_ = op.AddSystemLog(user, model.SystemLog{
		Type:    "info",
		Message: fmt.Sprintf("应用默认心跳脚本: %s", device.AndroidID),
		Source:  "device_script",
		IP:      c.ClientIP(),
	})
	common.SuccessResp(c)
}

type deleteDeviceReq struct {
	ID uint `json:"id" binding:"required"`
}

func DeleteDeviceScriptHandle(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req deleteDeviceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	device, err := op.GetDeviceByID(req.ID)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	ctx := op.BindHeartbeatUserToCtx(c.Request.Context())
	basePath := fmt.Sprintf("/sh/%s", device.AndroidID)
	if err := fs.Remove(ctx, basePath); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	// 同步删除设备记录
	_ = op.DeleteDevices([]uint{device.ID})
	_ = op.AddSystemLog(user, model.SystemLog{
		Type:    "info",
		Message: fmt.Sprintf("删除设备脚本: %s", device.AndroidID),
		Source:  "device_script",
		IP:      c.ClientIP(),
	})
	common.SuccessResp(c)
}

type deleteDevicesReq struct {
	IDs []uint `json:"ids" binding:"required"`
}

func DeleteDevices(c *gin.Context) {
	var req deleteDevicesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if err := op.DeleteDevices(req.IDs); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	common.SuccessResp(c)
}

type updateDeviceRemarkReq struct {
	ID     uint   `json:"id" binding:"required"`
	Remark string `json:"remark"`
}

func UpdateDeviceRemark(c *gin.Context) {
	var req updateDeviceRemarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if err := op.UpdateDeviceRemark(req.ID, req.Remark); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	common.SuccessResp(c)
}
