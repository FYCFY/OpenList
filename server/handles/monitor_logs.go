package handles

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type loginLogReq struct {
	Username string `json:"username"`
}

func LogLogin(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req loginLogReq
	_ = c.ShouldBindJSON(&req)
	if req.Username == "" {
		req.Username = user.Username
	}
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	if err := op.AddLoginLog(user, ip, ua); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"success": true,
		"data": gin.H{
			"username":   req.Username,
			"ip":         ip,
			"user_agent": ua,
		},
	})
}

type uploadLogReq struct {
	DeviceCode    string `json:"device_code"`
	SystemVersion string `json:"system_version"`
	FileName      string `json:"file_name"`
	FileSize      int64  `json:"file_size"`
}

func LogUpload(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req uploadLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	ip := c.ClientIP()
	payload := model.UploadLog{
		DeviceCode:    req.DeviceCode,
		SystemVersion: req.SystemVersion,
		FileName:      req.FileName,
		FileSize:      req.FileSize,
		IP:            ip,
	}
	if err := op.AddUploadLog(user, payload); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"success": true,
		"data":    payload,
	})
}

type systemLogReq struct {
	Type    string `json:"log_type" binding:"required"`
	Message string `json:"message" binding:"required"`
	Source  string `json:"source"`
	User    string `json:"user"`
}

func LogSystem(c *gin.Context) {
	user := c.Request.Context().Value(conf.UserKey).(*model.User)
	var req systemLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if req.Type != "info" && req.Type != "warning" && req.Type != "error" {
		common.ErrorStrResp(c, "Invalid log_type. Must be info, warning, or error", http.StatusBadRequest)
		return
	}
	ip := c.ClientIP()
	payload := model.SystemLog{
		Type:    req.Type,
		Message: req.Message,
		Source:  req.Source,
		IP:      ip,
	}
	if req.User != "" {
		payload.Username = req.User
	}
	if err := op.AddSystemLog(user, payload); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"success": true,
		"data":    payload,
	})
}

func parseRange(c *gin.Context) (start *time.Time, end *time.Time, err error) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	parse := func(v string) (*time.Time, error) {
		for _, layout := range layouts {
			if t, e := time.Parse(layout, v); e == nil {
				return &t, nil
			}
		}
		return nil, fmt.Errorf("invalid time: %s", v)
	}
	if startStr != "" {
		start, err = parse(startStr)
		if err != nil {
			return
		}
	}
	if endStr != "" {
		end, err = parse(endStr)
		if err != nil {
			return
		}
	}
	return
}

func listLogsInternal(c *gin.Context, defaultType string, clientShape bool) {
	logType := c.Query("type")
	if logType == "" {
		logType = defaultType
	}
	username := c.Query("username")
	start, end, err := parseRange(c)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	content, total, err := op.ListLogs(op.LogFilter{
		Type:     logType,
		Username: username,
		Start:    start,
		End:      end,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	if clientShape {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
			"success": true,
			"logs":    content,
			"total":   total,
		})
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: content,
		Total:   total,
	})
}

// Client-facing (Basic auth) list endpoints.
func ListLoginLogs(c *gin.Context)  { listLogsInternal(c, "login", true) }
func ListUploadLogs(c *gin.Context) { listLogsInternal(c, "upload", true) }
func ListSystemLogs(c *gin.Context) { listLogsInternal(c, "system", true) }

// Admin endpoints.
func AdminListLogs(c *gin.Context) { listLogsInternal(c, "login", false) }

type deleteLogsReq struct {
	Type string `json:"type"`
	IDs  []uint `json:"ids" binding:"required"`
}

func DeleteLogs(c *gin.Context) {
	var req deleteLogsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		req.Type = "login"
	}
	if err := op.DeleteLogs(req.Type, req.IDs); err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	common.SuccessResp(c)
}

func ExportLogs(c *gin.Context) {
	logType := c.Query("type")
	if logType == "" {
		logType = "login"
	}
	username := c.Query("username")
	start, end, err := parseRange(c)
	if err != nil {
		common.ErrorResp(c, err, http.StatusBadRequest)
		return
	}
	content, _, err := op.ListLogs(op.LogFilter{
		Type:     logType,
		Username: username,
		Start:    start,
		End:      end,
		Page:     1,
		PerPage:  10000,
	})
	if err != nil {
		common.ErrorResp(c, err, http.StatusInternalServerError, true)
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_logs.csv"`, logType))
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	switch logs := content.(type) {
	case []model.LoginLog:
		_ = writer.Write([]string{"ID", "Username", "IP", "UserAgent", "CreatedAt"})
		for _, l := range logs {
			_ = writer.Write([]string{
				strconv.FormatUint(uint64(l.ID), 10),
				l.Username,
				l.IP,
				l.UserAgent,
				l.CreatedAt.Format(time.RFC3339),
			})
		}
	case []model.UploadLog:
		_ = writer.Write([]string{"ID", "Username", "DeviceCode", "SystemVersion", "FileName", "FileSize", "IP", "CreatedAt"})
		for _, l := range logs {
			_ = writer.Write([]string{
				strconv.FormatUint(uint64(l.ID), 10),
				l.Username,
				l.DeviceCode,
				l.SystemVersion,
				l.FileName,
				strconv.FormatInt(l.FileSize, 10),
				l.IP,
				l.CreatedAt.Format(time.RFC3339),
			})
		}
	case []model.SystemLog:
		_ = writer.Write([]string{"ID", "Type", "Message", "Source", "Username", "IP", "CreatedAt"})
		for _, l := range logs {
			_ = writer.Write([]string{
				strconv.FormatUint(uint64(l.ID), 10),
				l.Type,
				l.Message,
				l.Source,
				l.Username,
				l.IP,
				l.CreatedAt.Format(time.RFC3339),
			})
		}
	default:
		log.Warnf("unknown log type for export: %T", content)
	}
}
