package handles

import (
	"errors"
	"time"

	"github.com/OpenListTeam/OpenList/v4/internal/davsession"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type listBlockReq struct {
	model.PageReq
}

func ListWebdavBlocks(c *gin.Context) {
	var req listBlockReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	blocks, total, err := davsession.ListBlocks(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: blocks,
		Total:   total,
	})
}

type addBlockReq struct {
	IP     string `json:"ip" binding:"required"`
	Remark string `json:"remark"`
	// Duration in number
	Duration int `json:"duration"`
	// Unit: minutes/hours/permanent
	Unit string `json:"unit" binding:"required"`
}

func parseExpire(duration int, unit string) (*time.Time, error) {
	switch unit {
	case "permanent":
		return nil, nil
	case "minutes":
		t := time.Now().Add(time.Duration(duration) * time.Minute)
		return &t, nil
	case "hours":
		t := time.Now().Add(time.Duration(duration) * time.Hour)
		return &t, nil
	default:
		return nil, errors.New("invalid unit")
	}
}

func AddWebdavBlock(c *gin.Context) {
	var req addBlockReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	expireAt, err := parseExpire(req.Duration, req.Unit)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := davsession.AddBlock(req.IP, req.Remark, expireAt); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

type updateBlockReq struct {
	ID       uint   `json:"id" binding:"required"`
	Remark   string `json:"remark"`
	Duration int    `json:"duration"`
	Unit     string `json:"unit" binding:"required"`
}

func UpdateWebdavBlock(c *gin.Context) {
	var req updateBlockReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	expireAt, err := parseExpire(req.Duration, req.Unit)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := davsession.UpdateBlock(req.ID, req.Remark, expireAt); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

type deleteBlockReq struct {
	ID uint `json:"id" binding:"required"`
}

func DeleteWebdavBlock(c *gin.Context) {
	var req deleteBlockReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := davsession.DeleteBlock(req.ID); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
