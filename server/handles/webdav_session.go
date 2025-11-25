package handles

import (
	"github.com/OpenListTeam/OpenList/v4/internal/davsession"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

type ListWebdavSessionReq struct {
	model.PageReq
	Username string `json:"username" form:"username"`
}

func ListWebdavSessions(c *gin.Context) {
	var req ListWebdavSessionReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	sessions, total, err := davsession.ListSessions(req.Username, req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: sessions,
		Total:   total,
	})
}

type DisconnectWebdavSessionReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

func DisconnectWebdavSession(c *gin.Context) {
	var req DisconnectWebdavSessionReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := davsession.ForceCloseSession(req.ID); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
