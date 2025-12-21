package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func EmailQueueManage(c *gin.Context) {
	stats, err := getEmailQueueStats()
	if err != nil {
		common.HandleMessage(c, "获取队列状态失败: "+err.Error())
		return
	}

	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "email_queue.html", gin.H{
		"user":     user,
		"cfg":      system.GetConfiguration(),
		"stats":    stats,
		"title":    "邮件队列管理",
		"comments": models.MustListUnreadComment(),
	})
}
