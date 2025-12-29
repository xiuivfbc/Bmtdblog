package content

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 切换页面发布状态
// @Description 切换指定ID页面的发布状态（发布/下架）
// @Tags 页面管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "页面ID"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/page/{id}/publish [post]
func PagePublish(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	page, err := models.GetPageById(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	page.IsPublished = !page.IsPublished
	err = page.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
