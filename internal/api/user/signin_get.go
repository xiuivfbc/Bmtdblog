package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 用户登录页面
// @Description 显示用户登录表单页面
// @Tags 用户认证
// @Accept html
// @Produce html
// @Success 200 {html} string "登录页面"
// @Router /signin [get]
func SigninGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signin.html", gin.H{
		"cfg": config.GetConfiguration(),
	})
}
