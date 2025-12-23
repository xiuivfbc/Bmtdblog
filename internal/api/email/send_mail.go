package email

import (
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
