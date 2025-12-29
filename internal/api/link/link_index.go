package link

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 友情链接列表
// @Description 显示管理后台的友情链接列表
// @Tags 友情链接管理
// @Accept html
// @Produce html
// @Success 200 {html} string "友情链接列表页面"
// @Router /admin/link [get]
func LinkIndex(c *gin.Context) {
	links, _ := models.ListLinks()
	c.HTML(http.StatusOK, "admin/link.html", gin.H{
		"links":    links,
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      config.GetConfiguration(),
	})
}
