package queue

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 邮件队列管理
// @Description 查看邮件队列状态和管理邮件队列
// @Tags 队列管理
// @Accept json
// @Produce html
// @Success 200 {string} html "邮件队列管理页面"
// @Failure 200 {string} string "获取队列状态失败的消息"
// @Router /admin/email_queue [get]
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
