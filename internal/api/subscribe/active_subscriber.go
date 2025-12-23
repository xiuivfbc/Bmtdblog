package subscribe

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func ActiveSubscriber(c *gin.Context) {
	var (
		err        error
		subscriber *models.Subscriber
	)
	sid := c.Query("sid")
	if sid == "" {
		common.HandleMessage(c, "激活链接有误，请重新获取！")
		return
	}
	subscriber, err = models.GetSubscriberBySignature(sid)
	if err != nil {
		common.HandleMessage(c, "激活链接有误，请重新获取！")
		return
	}
	if !common.GetCurrentTime().Before(subscriber.OutTime) {
		common.HandleMessage(c, "激活链接已过期，请重新获取！")
		return
	}
	subscriber.VerifyState = true
	subscriber.OutTime = common.GetCurrentTime()
	err = subscriber.Update()
	if err != nil {
		common.HandleMessage(c, fmt.Sprintf("激活失败！%s", err.Error()))
		return
	}
	common.HandleMessage(c, "激活成功！")
}
