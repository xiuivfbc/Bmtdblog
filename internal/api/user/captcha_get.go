package user

import (
	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

// @Summary 获取验证码
// @Description 生成并返回验证码图片
// @Tags 用户认证
// @Produce image/png
// @Success 200 {file} binary "验证码图片"
// @Router /captcha [get]
func CaptchaGet(context *gin.Context) {
	session := sessions.Default(context)
	captchaId := captcha.NewLen(4)
	session.Delete(common.SessionCaptcha)
	session.Set(common.SessionCaptcha, captchaId)
	session.Save()
	captcha.WriteImage(context.Writer, captchaId, 100, 40)
}
