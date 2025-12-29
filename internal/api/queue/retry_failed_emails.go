package queue

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
)

// @Summary 重试失败邮件
// @Description 重新发送队列中失败的邮件
// @Tags 队列管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "{"success":true,"message":"重试成功","count":10}"
// @Failure 500 {object} map[string]interface{} "{"success":false,"message":"错误信息"}"
// @Router /admin/queue/retry_failed_emails [post]
func RetryFailedEmails(c *gin.Context) {
	log.Debug("RetryFailedEmails")
	count, err := retryFailedEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "重试成功",
		"count":   count,
	})
}

// 重试失败的邮件
func retryFailedEmails() (int, error) {
	log.Debug("retryFailedEmails")
	if dao.EmailQueueInstance == nil {
		return 0, fmt.Errorf("邮件队列未启用")
	}
	return dao.EmailQueueInstance.RetryFailedEmails()
}
