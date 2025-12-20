package system

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func SendMail(c *gin.Context) {
	var (
		err        error
		res        = gin.H{}
		uid        uint
		subscriber *models.Subscriber
	)
	defer common.WriteJSON(c, res)
	subject := c.PostForm("subject")
	content := c.PostForm("content")
	userId := c.Query("userId")

	if subject == "" || content == "" || userId == "" {
		res["message"] = "error parameter"
		return
	}
	uid, err = common.ParseUint(userId)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	subscriber, err = models.GetSubscriberById(uid)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	err = common.SendMail(subscriber.Email, subject, content)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func SendBatchMail(c *gin.Context) {
	var (
		err         error
		res         = gin.H{}
		subscribers []*models.Subscriber
		emails      = make([]string, 0)
	)
	defer common.WriteJSON(c, res)
	subject := c.PostForm("subject")
	content := c.PostForm("content")
	if subject == "" || content == "" {
		res["message"] = "error parameter"
		return
	}
	subscribers, err = models.ListSubscriber(true)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	for _, subscriber := range subscribers {
		emails = append(emails, subscriber.Email)
	}
	err = common.SendMail(strings.Join(emails, ";"), subject, content)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
