package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 用户列表
// @Description 显示管理后台的用户列表
// @Tags 用户管理
// @Accept html
// @Produce html
// @Success 200 {html} string "用户列表页面"
// @Router /admin/user [get]
func UserIndex(c *gin.Context) {
	users, _ := models.ListUsers()
	c.HTML(http.StatusOK, "admin/user.html", gin.H{
		"users":    users,
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      config.GetConfiguration(),
	})
}
