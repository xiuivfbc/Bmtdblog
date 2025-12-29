package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 创建文章表单
// @Description 管理后台创建新文章的表单页面
// @Tags 文章管理
// @Accept html
// @Produce html
// @Success 200 {html} string "创建文章表单"
// @Router /admin/new_post [get]
func PostNew(c *gin.Context) {
	c.HTML(http.StatusOK, "post/new.html", gin.H{
		"user": c.MustGet(common.ContextUserKey),
		"cfg":  config.GetConfiguration(),
	})
}
