package subscribe

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 激活订阅者
// @Description 通过激活链接激活订阅者邮箱
// @Tags 订阅管理
// @Accept html
// @Produce html
// @Param sid query string true "激活签名"
// @Success 200 {string} string "激活成功或失败的消息"
// @Failure 200 {string} string "激活失败的消息"
// @Router /active [get]
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
