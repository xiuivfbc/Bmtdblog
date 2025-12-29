package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 管理后台首页
// @Description 显示管理后台的统计信息和仪表盘
// @Tags 管理后台
// @Accept html
// @Produce html
// @Success 200 {html} string "管理后台首页"
// @Router /admin/index [get]
func AdminIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"pageCount":    models.CountPage(),
		"postCount":    models.CountPost(),
		"tagCount":     models.CountTag(),
		"commentCount": models.CountComment(),
		"user":         c.MustGet(common.ContextUserKey),
		"comments":     models.MustListUnreadComment(),
		"cfg":          config.GetConfiguration(),
	})
}
