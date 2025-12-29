package user

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
)

// @Summary 用户登出
// @Description 退出当前用户登录状态
// @Tags 用户认证
// @Accept json
// @Produce html
// @Success 302 {string} string "登出成功，重定向到首页"
// @Router /logout [get]
func LogoutGet(c *gin.Context) {
	log.Debug("LogoutGet")
	s := sessions.Default(c)
	s.Clear()
	s.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
