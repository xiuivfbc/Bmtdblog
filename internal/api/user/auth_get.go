package user

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"go.uber.org/zap"
)

// @Summary 获取认证URL
// @Description 根据认证类型获取第三方登录URL
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param authType path string true "认证类型，如github"
// @Success 302 {string} string "重定向到认证页面"
// @Failure 400 {string} string "参数错误"
// @Router /auth/{authType} [get]
func AuthGet(c *gin.Context) {
	authType := c.Param("authType")
	log.Debug("AuthGet", zap.String("authType", authType))

	session := sessions.Default(c)
	uuid := common.UUID()
	session.Delete(common.SessionGithubState)
	session.Set(common.SessionGithubState, uuid)
	session.Save()

	cfg := config.GetConfiguration()

	authurl := "/signin"
	switch authType {
	case "github":
		authurl = fmt.Sprintf(cfg.Github.AuthUrl, cfg.Github.ClientId, uuid)
	default:
	}
	c.Redirect(http.StatusFound, authurl)
}
