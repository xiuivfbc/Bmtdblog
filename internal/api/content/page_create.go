package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 创建新页面
// @Description 创建新的博客页面
// @Tags 页面管理
// @Accept x-www-form-urlencoded
// @Produce html
// @Param title formData string true "页面标题"
// @Param body formData string true "页面内容"
// @Param isPublished formData string false "是否发布(on表示发布)"
// @Success 301 {string} redirect "重定向到页面管理页面"
// @Failure 200 {string} html "创建失败的页面"
// @Router /admin/page/create [POST]
func PageCreate(c *gin.Context) {
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"
	log.Debug("PageCreate", zap.String("title", title), zap.Bool("isPublished", published))

	page := &models.Page{
		Title:       title,
		Body:        body,
		IsPublished: published,
	}
	err := page.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "page/new.html", gin.H{
			"message": err.Error(),
			"page":    page,
			"user":    c.MustGet(common.ContextUserKey),
			"cfg":     config.GetConfiguration(),
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/page")
}
