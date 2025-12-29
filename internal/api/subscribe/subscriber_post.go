package subscribe

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"go.uber.org/zap"
)

// @Summary 发送订阅邮件
// @Description 向订阅者发送邮件，邮箱为空时发送给所有订阅者
// @Tags 订阅者管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param mail formData string false "邮件地址，为空时发送给所有订阅者"
// @Param subject formData string true "邮件主题"
// @Param body formData string true "邮件内容"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/subscriber [post]
func SubscriberPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	mail := c.PostForm("mail")
	subject := c.PostForm("subject")
	body := c.PostForm("body")
	log.Debug("SubscriberPost", zap.String("mail", mail), zap.String("subject", subject))
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
