package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 搜索首页
// @Description 显示搜索表单页面
// @Tags 搜索功能
// @Accept html
// @Produce html
// @Success 200 {html} string "搜索首页"
// @Router /search/index [get]
func SearchIndexGet(c *gin.Context) {
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "search/index.html", gin.H{
		"user":    user,
		"allTags": models.MustListTag(),
		"cfg":     config.GetConfiguration(),
	})
}
