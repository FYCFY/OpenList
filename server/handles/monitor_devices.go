package handles

import (
	"net/http"
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
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
