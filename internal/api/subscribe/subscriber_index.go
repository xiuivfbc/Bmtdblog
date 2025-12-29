package subscribe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 订阅者列表
// @Description 显示管理后台的订阅者列表
// @Tags 订阅者管理
// @Accept html
// @Produce html
// @Success 200 {html} string "订阅者列表页面"
// @Router /admin/subscriber [get]
func SubscriberIndex(c *gin.Context) {
	subscribers, _ := models.ListSubscriber(false)
	c.HTML(http.StatusOK, "admin/subscriber.html", gin.H{
		"subscribers": subscribers,
		"user":        c.MustGet(common.ContextUserKey),
		"comments":    models.MustListUnreadComment(),
		"cfg":         config.GetConfiguration(),
	})
}
