package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

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
