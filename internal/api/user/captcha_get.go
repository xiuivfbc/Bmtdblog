package user

import (
	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

func CaptchaGet(context *gin.Context) {
	session := sessions.Default(context)
	captchaId := captcha.NewLen(4)
	session.Delete(common.SessionCaptcha)
	session.Set(common.SessionCaptcha, captchaId)
	session.Save()
	captcha.WriteImage(context.Writer, captchaId, 100, 40)
}
