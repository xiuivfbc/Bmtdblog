package content

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 创建标签
// @Description 创建新的文章标签
// @Tags 标签管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param value formData string true "标签名称"
// @Success 200 {object} map[string]interface{} "{"succeed":true,"data":{"id":1,"name":"标签名"}}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/tag/create [post]
func TagCreate(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	name := c.PostForm("value")
	tag := &models.Tag{Name: name}
	err = tag.Insert()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["data"] = tag
}
