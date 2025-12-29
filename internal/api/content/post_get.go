package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 获取文章详情
// @Description 根据ID获取文章的详细内容
// @Tags 内容管理
// @Accept json
// @Produce html
// @Param id path uint true "文章ID"
// @Success 200 {string} string "文章详情页面"
// @Failure 404 {string} string "文章不存在或未发布"
// @Router /post/{id} [get]
func PostGet(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostByIdWithCache(id)
	if err != nil || !post.IsPublished {
		common.Handle404(c)
		return
	}
	// 更新浏览数（异步，避免影响缓存和性能）
	go func() {
		post.View++
		post.UpdateView()
	}()

	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "post/display.html", gin.H{
		"post": post,
		"user": user,
		"cfg":  config.GetConfiguration(),
	})
}
