package interaction

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func UnSubscribe(c *gin.Context) {
	fmt.Println("UnSubscribe")
	userId := c.Query("userId")
	if userId == "" {
		common.HandleMessage(c, "Internal Server Error!")
		return
	}
	temp, _ := strconv.Atoi(userId)
	userID := uint(temp)
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
