package subscribe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func SubscribeGet(c *gin.Context) {
	count, _ := models.CountSubscriber()
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
		"total": count,
		"user":  user,
		"cfg":   config.GetConfiguration(),
	})
}
