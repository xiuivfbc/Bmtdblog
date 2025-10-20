package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
)

// EmailQueueStatus 邮件队列状态页面
func EmailQueueStatus(c *gin.Context) {
	stats, err := getEmailQueueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// EmailQueueManage 邮件队列管理页面
func EmailQueueManage(c *gin.Context) {
	stats, err := getEmailQueueStats()
	if err != nil {
		HandleMessage(c, "获取队列状态失败: "+err.Error())
		return
	}

	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusOK, "email_queue.html", gin.H{
		"user":     user,
		"cfg":      system.GetConfiguration(),
		"stats":    stats,
		"title":    "邮件队列管理",
		"comments": models.MustListUnreadComment(),
	})
}

// RetryFailedEmails 重试失败的邮件
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

// ClearFailedEmails 清理失败队列
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

// 获取邮件队列统计
func getEmailQueueStats() (map[string]interface{}, error) {
	if system.EmailQueueInstance == nil {
		return map[string]interface{}{
			"status":      "disabled",
			"workers":     0,
			"queue_size":  0,
			"failed_size": 0,
		}, nil
	}

	return system.EmailQueueInstance.GetQueueStats()
}

// 重试失败的邮件
func retryFailedEmails() (int, error) {
	if system.EmailQueueInstance == nil {
		return 0, fmt.Errorf("邮件队列未启用")
	}
	return system.EmailQueueInstance.RetryFailedEmails()
}

// 清理失败队列
func clearFailedEmails() (int, error) {
	if system.EmailQueueInstance == nil {
		return 0, fmt.Errorf("邮件队列未启用")
	}
	return system.EmailQueueInstance.ClearFailedEmails()
}
