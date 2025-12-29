package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 编辑文章表单
// @Description 管理后台编辑文章的表单页面
// @Tags 文章管理
// @Accept html
// @Produce html
// @Param id path uint true "文章ID"
// @Success 200 {html} string "编辑文章表单"
// @Failure 404 {string} string "文章不存在"
// @Router /admin/post/{id}/edit [get]
func PostEdit(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostByIdWithCache(id)
	if err != nil {
		common.Handle404(c)
		return
	}
	c.HTML(http.StatusOK, "post/modify.html", gin.H{
		"post": post,
		"user": c.MustGet(common.ContextUserKey),
		"cfg":  config.GetConfiguration(),
	})
}
