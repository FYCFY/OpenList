package middlewares

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/OpenListTeam/OpenList/v4/internal/conf"
	"github.com/OpenListTeam/OpenList/v4/internal/model"
	"github.com/OpenListTeam/OpenList/v4/internal/op"
	"github.com/OpenListTeam/OpenList/v4/internal/setting"
	"github.com/OpenListTeam/OpenList/v4/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// WebdavBasicAPI authenticates requests using WebDAV credentials via BasicAuth.
// It mirrors the WebDAV auth checks but responds with JSON.
func WebdavBasicAPI(c *gin.Context) {
	ip := c.ClientIP()
	count, cok := model.LoginCache.Get(ip)
	if cok && count >= model.DefaultMaxAuthRetries {
		common.ErrorStrResp(c, "Too many unsuccessful sign-in attempts, try again later.", http.StatusTooManyRequests)
		model.LoginCache.Expire(ip, model.DefaultLockDuration)
		return
	}

	username, password, ok := c.Request.BasicAuth()
	if !ok {
		bt := c.GetHeader("Authorization")
		if strings.HasPrefix(bt, "Bearer") {
			bt = strings.TrimPrefix(bt, "Bearer ")
			token := setting.GetStr(conf.Token)
			if token != "" && subtle.ConstantTimeCompare([]byte(bt), []byte(token)) == 1 {
				admin, err := op.GetAdmin()
				if err != nil {
					common.ErrorResp(c, err, http.StatusInternalServerError)
					return
				}
				common.GinWithValue(c, conf.UserKey, admin)
				c.Next()
				return
			}
		}
		common.ErrorStrResp(c, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := op.GetUserByName(username)
	if err != nil || user.ValidateRawPassword(password) != nil {
		model.LoginCache.Set(ip, count+1)
		common.ErrorStrResp(c, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// at least auth is successful till here
	model.LoginCache.Del(ip)
	if user.Disabled || !user.CanWebdavRead() {
		common.ErrorStrResp(c, "Forbidden", http.StatusForbidden)
		return
	}

	common.GinWithValue(c, conf.UserKey, user)
	log.Debugf("use basic webdav auth: %+v", user)
	c.Next()
}
