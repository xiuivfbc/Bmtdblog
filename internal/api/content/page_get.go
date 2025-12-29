package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 页面详情
// @Description 根据ID获取静态页面的详细内容
// @Tags 内容浏览
// @Accept html
// @Produce html
// @Param id path uint true "页面ID"
// @Success 200 {html} string "页面详情页面"
// @Failure 404 {string} string "页面不存在或未发布"
// @Router /page/{id} [get]
func PageGet(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	page, err := models.GetPageById(id)
	if err != nil || !page.IsPublished {
		common.Handle404(c)
		return
	}
	page.View++
	page.UpdateView()
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "page/display.html", gin.H{
		"page": page,
		"user": user,
		"cfg":  config.GetConfiguration(),
	})
}
