package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 创建页面表单
// @Description 管理后台创建新页面的表单页面
// @Tags 页面管理
// @Accept html
// @Produce html
// @Success 200 {html} string "创建页面表单"
// @Router /admin/new_page [get]
func PageNew(c *gin.Context) {
	c.HTML(http.StatusOK, "page/new.html", gin.H{
		"user": c.MustGet(common.ContextUserKey),
		"cfg":  config.GetConfiguration(),
	})
}
