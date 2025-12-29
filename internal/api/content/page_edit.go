package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 编辑页面表单
// @Description 管理后台编辑页面的表单页面
// @Tags 页面管理
// @Accept html
// @Produce html
// @Param id path uint true "页面ID"
// @Success 200 {html} string "编辑页面表单"
// @Failure 404 {string} string "页面不存在"
// @Router /admin/page/{id}/edit [get]
func PageEdit(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	page, err := models.GetPageById(id)
	if err != nil {
		common.Handle404(c)
		return
	}
	c.HTML(http.StatusOK, "page/modify.html", gin.H{
		"page": page,
		"user": c.MustGet(common.ContextUserKey),
		"cfg":  config.GetConfiguration(),
	})
}
