package system

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func RetryFailedEmails(c *gin.Context) {
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
	if system.EmailQueueInstance == nil {
		return 0, fmt.Errorf("邮件队列未启用")
	}
	return system.EmailQueueInstance.RetryFailedEmails()
}
