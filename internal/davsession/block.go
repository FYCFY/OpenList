package davsession

import (
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/db"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"gorm.io/gorm"
)

func cleanExpiredBlocks(now time.Time, d *gorm.DB) {
	_ = d.Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&model.WebdavBlock{}).Error
}

// IsBlocked checks if the ip is currently blocked for WebDAV.
func IsBlocked(ip string) bool {
	d := db.GetDb()
	now := time.Now()
	cleanExpiredBlocks(now, d)
	var block model.WebdavBlock
	err := d.Where("ip = ?", ip).First(&block).Error
	if err != nil {
		return false
	}
	if block.ExpiresAt != nil && block.ExpiresAt.Before(now) {
		_ = d.Delete(&block).Error
		return false
	}
	return true
}

// AddBlock creates/updates a block for an IP with optional expire time.
func AddBlock(ip, remark string, expireAt *time.Time) error {
	d := db.GetDb()
	now := time.Now()
	cleanExpiredBlocks(now, d)
	var block model.WebdavBlock
	err := d.Where("ip = ?", ip).First(&block).Error
	if err == nil {
		return d.Model(&block).Updates(map[string]interface{}{
			"remark":     remark,
			"expires_at": expireAt,
		}).Error
	}
	block = model.WebdavBlock{
		IP:        ip,
		Remark:    remark,
		ExpiresAt: expireAt,
	}
	return d.Create(&block).Error
}

// UpdateBlock updates an existing block by id.
func UpdateBlock(id uint, remark string, expireAt *time.Time) error {
	d := db.GetDb()
	now := time.Now()
	cleanExpiredBlocks(now, d)
	return d.Model(&model.WebdavBlock{}).Where("id = ?", id).Updates(map[string]interface{}{
		"remark":     remark,
		"expires_at": expireAt,
	}).Error
}

// DeleteBlock removes a block by id.
func DeleteBlock(id uint) error {
	return db.GetDb().Delete(&model.WebdavBlock{}, id).Error
}

// ListBlocks returns paginated block records.
func ListBlocks(page, perPage int) ([]model.WebdavBlock, int64, error) {
	d := db.GetDb()
	now := time.Now()
	cleanExpiredBlocks(now, d)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	query := d.Model(&model.WebdavBlock{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var blocks []model.WebdavBlock
	if err := query.Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&blocks).Error; err != nil {
		return nil, 0, err
	}
	return blocks, total, nil
}
