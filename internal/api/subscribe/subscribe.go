package subscribe

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func Subscribe(c *gin.Context) {
	mail := c.PostForm("mail")
	user, _ := c.Get(common.ContextUserKey)
	log.Debug("Subscribe", zap.String("mail", mail))
	var err error
	if len(mail) > 0 {
		var subscriber *models.Subscriber
		subscriber, err = models.GetSubscriberByEmail(mail)
		if err == nil {
			// 已存在
			if !subscriber.VerifyState && common.GetCurrentTime().After(subscriber.OutTime) {
				// 未激活或激活过期
				err = sendActiveEmail(subscriber)
				if err == nil {
					count, _ := models.CountSubscriber()
					c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
						"message": "subscribe succeed.",
						"total":   count,
						"user":    user,
						"cfg":     config.GetConfiguration(),
					})
					return
				}
			} else if subscriber.VerifyState && !subscriber.SubscribeState {
				// 已激活但未订阅
				subscriber.SubscribeState = true
				err = subscriber.Update()
				if err == nil {
					err = errors.New("subscribe succeed.")
				}
			} else {
				err = errors.New("mail have already actived or have unactive mail in your mailbox.")
			}
		} else {
			// 不存在
			subscriber := &models.Subscriber{
				Email: mail,
			}
			err = subscriber.Insert()
			if err == nil {
				err = sendActiveEmail(subscriber)
				if err == nil {
					count, _ := models.CountSubscriber()
					c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
						"message": "subscribe succeed.",
						"total":   count,
						"user":    user,
						"cfg":     config.GetConfiguration(),
					})
					return
				}
			}
		}
	} else {
		// 邮箱为空
		err = errors.New("empty mail address.")
	}
	count, _ := models.CountSubscriber()
	c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
		"message": err.Error(),
		"total":   count,
		"user":    user,
		"cfg":     config.GetConfiguration(),
	})
}
