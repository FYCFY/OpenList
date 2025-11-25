package handles

import (
	"strconv"

	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
)

func ClearWebdavBind(c *gin.Context) {
	idStr := c.Query("user_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := op.ClearUserBindIP(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
