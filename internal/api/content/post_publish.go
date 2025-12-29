package content

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 切换文章发布状态
// @Description 切换指定ID文章的发布状态（发布/下架）
// @Tags 文章管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "文章ID"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/post/{id}/publish [post]
func PostPublish(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
	)
	defer common.WriteJSON(c, res)
	id, err := common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post, err = models.GetPostByIdWithCache(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post.IsPublished = !post.IsPublished
	err = post.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
