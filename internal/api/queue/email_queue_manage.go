package queue

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func EmailQueueManage(c *gin.Context) {
	log.Debug("EmailQueueManage")
	stats, err := getEmailQueueStats()
	if err != nil {
		common.HandleMessage(c, "获取队列状态失败: "+err.Error())
		return
	}

	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "email_queue.html", gin.H{
		"user":     user,
		"cfg":      config.GetConfiguration(),
		"stats":    stats,
		"title":    "邮件队列管理",
		"comments": models.MustListUnreadComment(),
	})
}
