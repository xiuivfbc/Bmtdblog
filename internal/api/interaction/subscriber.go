package interaction

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func SubscribeGet(c *gin.Context) {
	count, _ := models.CountSubscriber()
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "other/subscribe.html", gin.H{
		"total": count,
		"user":  user,
		"cfg":   system.GetConfiguration(),
	})
}

func Subscribe(c *gin.Context) {
	mail := c.PostForm("mail")
	user, _ := c.Get(common.ContextUserKey)
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
						"cfg":     system.GetConfiguration(),
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
						"cfg":     system.GetConfiguration(),
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
		"cfg":     system.GetConfiguration(),
	})
}

func sendActiveEmail(subscriber *models.Subscriber) (err error) {
	uuid := common.UUID()
	duration, _ := time.ParseDuration("30m")
	subscriber.OutTime = common.GetCurrentTime().Add(duration)
	subscriber.SecretKey = uuid
	signature := common.Md5(subscriber.Email + uuid + subscriber.OutTime.Format("20060102150405"))
	subscriber.Signature = signature
	err = common.SendMail(subscriber.Email, fmt.Sprintf("[%s]邮箱验证", system.GetConfiguration().Title), fmt.Sprintf("%s/active?sid=%s", system.GetConfiguration().Domain, signature))
	if err != nil {
		return
	}
	err = subscriber.Update()
	return
}

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

func sendEmailToSubscribers(subject, body string) (err error) {
	var (
		subscribers []*models.Subscriber
		emails      = make([]string, 0)
	)
	subscribers, err = models.ListSubscriber(true)
	if err != nil {
		return
	}
	for _, subscriber := range subscribers {
		emails = append(emails, subscriber.Email)
	}
	if len(emails) == 0 {
		err = errors.New("no subscribers!")
		return
	}
	err = common.SendMail(strings.Join(emails, ";"), subject, body)
	return
}

func SubscriberIndex(c *gin.Context) {
	subscribers, _ := models.ListSubscriber(false)
	c.HTML(http.StatusOK, "admin/subscriber.html", gin.H{
		"subscribers": subscribers,
		"user":        c.MustGet(common.ContextUserKey),
		"comments":    models.MustListUnreadComment(),
		"cfg":         system.GetConfiguration(),
	})
}

// 邮箱为空时，发送给所有订阅者
func SubscriberPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	mail := c.PostForm("mail")
	subject := c.PostForm("subject")
	body := c.PostForm("body")
	if len(mail) > 0 {
		err = common.SendMail(mail, subject, body)
	} else {
		err = sendEmailToSubscribers(subject, body)
	}
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
