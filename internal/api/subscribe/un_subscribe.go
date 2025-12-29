package subscribe

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 取消订阅
// @Description 取消博客订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param userId query string true "订阅用户ID"
// @Success 200 {object} map[string]interface{} "{"msg":"Unsubscribe Successful!","succeed":true}"
// @Failure 200 {string} string "取消订阅失败的消息"
// @Router /unsubscribe [get]
func UnSubscribe(c *gin.Context) {
	fmt.Println("UnSubscribe")
	userId := c.Query("userId")
	if userId == "" {
		common.HandleMessage(c, "Internal Server Error!")
		return
	}
	temp, _ := strconv.Atoi(userId)
	userID := uint(temp)
	log.Debug("UnSubscribe", zap.Uint("userID", userID))
	subscriber, err := models.GetSubscriberById(userID)
	if err != nil || !subscriber.VerifyState || !subscriber.SubscribeState {
		common.HandleMessage(c, "Unscribe failed.")
		return
	}
	subscriber.SubscribeState = false
	err = subscriber.Update()
	if err != nil {
		common.HandleMessage(c, fmt.Sprintf("Unscribe failed.%s", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":     "Unsubscribe Successful!",
		"succeed": true,
	})
}
