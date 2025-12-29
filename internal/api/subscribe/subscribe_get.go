package subscribe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 订阅页面
// @Description 显示博客订阅页面
// @Tags 订阅管理
// @Accept html
// @Produce html
// @Success 200 {html} string "订阅页面"
// @Router /subscribe [get]
func SubscribeGet(c *gin.Context) {
	count, _ := models.CountSubscriber()
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
		"total": count,
		"user":  user,
		"cfg":   config.GetConfiguration(),
	})
}
