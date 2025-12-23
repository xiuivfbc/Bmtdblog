package subscribe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func SubscriberIndex(c *gin.Context) {
	subscribers, _ := models.ListSubscriber(false)
	c.HTML(http.StatusOK, "admin/subscriber.html", gin.H{
		"subscribers": subscribers,
		"user":        c.MustGet(common.ContextUserKey),
		"comments":    models.MustListUnreadComment(),
		"cfg":         config.GetConfiguration(),
	})
}
