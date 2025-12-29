package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 获取个人资料
// @Description 显示管理后台的个人资料页面
// @Tags 个人资料管理
// @Accept html
// @Produce html
// @Success 200 {html} string "个人资料页面"
// @Router /admin/profile [get]
func ProfileGet(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/profile.html", gin.H{
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      config.GetConfiguration(),
	})
}
