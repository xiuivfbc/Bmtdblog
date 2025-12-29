package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 用户注册页面
// @Description 显示用户注册表单页面
// @Tags 用户认证
// @Accept html
// @Produce html
// @Success 200 {html} string "注册页面"
// @Router /signup [get]
func SignupGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signup.html", gin.H{
		"cfg": config.GetConfiguration(),
	})
}
