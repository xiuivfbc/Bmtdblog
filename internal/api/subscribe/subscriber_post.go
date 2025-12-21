package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

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
