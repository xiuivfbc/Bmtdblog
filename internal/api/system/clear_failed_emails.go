package system

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func ClearFailedEmails(c *gin.Context) {
	count, err := clearFailedEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "清理成功",
		"count":   count,
	})
}

// 清理失败队列
func clearFailedEmails() (int, error) {
	if system.EmailQueueInstance == nil {
		return 0, fmt.Errorf("邮件队列未启用")
	}
	return system.EmailQueueInstance.ClearFailedEmails()
}
