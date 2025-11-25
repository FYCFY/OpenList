package op

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
)

func AddLoginLog(user *model.User, ip, ua string) error {
	now := time.Now()
	var userID *uint
	if user != nil {
		userID = &user.ID
	}
	return db.GetDb().Create(&model.LoginLog{
		UserID:    userID,
		Username:  user.Username,
		IP:        ip,
		UserAgent: ua,
		CreatedAt: now,
	}).Error
}

type LogFilter struct {
	Type     string
	Username string
	Start    *time.Time
	End      *time.Time
	Page     int
	PerPage  int
}

func AddUploadLog(user *model.User, payload model.UploadLog) error {
	payload.UserID = &user.ID
	payload.Username = user.Username
	payload.CreatedAt = time.Now()
	return db.GetDb().Create(&payload).Error
}

func AddSystemLog(user *model.User, payload model.SystemLog) error {
	payload.UserID = &user.ID
	if payload.Username == "" {
		payload.Username = user.Username
	}
	payload.CreatedAt = time.Now()
	return db.GetDb().Create(&payload).Error
}

func ListLogs(filter LogFilter) (interface{}, int64, error) {
	switch filter.Type {
	case "upload":
		return listUploadLogs(filter)
	case "system":
		return listSystemLogs(filter)
	default:
		return listLoginLogs(filter)
	}
}

func listLoginLogs(filter LogFilter) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	query := db.GetDb().Model(&model.LoginLog{})
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	if filter.Start != nil {
		query = query.Where("created_at >= ?", *filter.Start)
	}
	if filter.End != nil {
		query = query.Where("created_at <= ?", *filter.End)
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
	if err := query.Order("created_at DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func listUploadLogs(filter LogFilter) ([]model.UploadLog, int64, error) {
	var logs []model.UploadLog
	query := db.GetDb().Model(&model.UploadLog{})
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	if filter.Start != nil {
		query = query.Where("created_at >= ?", *filter.Start)
	}
	if filter.End != nil {
		query = query.Where("created_at <= ?", *filter.End)
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
	if err := query.Order("created_at DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func listSystemLogs(filter LogFilter) ([]model.SystemLog, int64, error) {
	var logs []model.SystemLog
	query := db.GetDb().Model(&model.SystemLog{})
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	if filter.Start != nil {
		query = query.Where("created_at >= ?", *filter.Start)
	}
	if filter.End != nil {
		query = query.Where("created_at <= ?", *filter.End)
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
	if err := query.Order("created_at DESC").Limit(perPage).Offset((page - 1) * perPage).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func DeleteLogs(kind string, ids []uint) error {
	var modelType interface{}
	switch kind {
	case "upload":
		modelType = &model.UploadLog{}
	case "system":
		modelType = &model.SystemLog{}
	default:
		modelType = &model.LoginLog{}
	}
	return db.GetDb().Where("id IN ?", ids).Delete(modelType).Error
}
