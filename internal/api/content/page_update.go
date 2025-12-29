package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 更新页面
// @Description 更新指定ID的页面内容
// @Tags 页面管理
// @Accept x-www-form-urlencoded
// @Produce html
// @Param id path uint true "页面ID"
// @Param title formData string true "页面标题"
// @Param body formData string true "页面内容"
// @Param isPublished formData string false "是否发布(on表示发布)"
// @Success 301 {string} string "更新成功，重定向到页面列表"
// @Failure 500 {string} string "服务器内部错误"
// @Router /admin/page/{id}/edit [post]
func PageUpdate(c *gin.Context) {
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"

	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	log.Debug("PageUpdate", zap.Uint("id", id), zap.String("title", title), zap.Bool("isPublished", published))
	page := &models.Page{Title: title, Body: body, IsPublished: published}
	page.ID = id
	err = page.Update()
	if err != nil {
		log.Error("page.Update error", "err", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/page")
}
