package user

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 用户注册提交
// @Description 处理用户注册表单提交，创建新用户
// @Tags 用户认证
// @Accept multipart/form-data
// @Produce html
// @Param email formData string true "邮箱地址"
// @Param telephone formData string false "手机号码"
// @Param password formData string true "密码"
// @Param verifyCode formData string true "验证码"
// @Success 301 {string} string "注册成功，重定向到登录页"
// @Failure 200 {html} string "注册失败，返回注册页面并显示错误信息"
// @Router /signup [post]
func SignupPost(c *gin.Context) {
	var (
		err error
	)
	email := c.PostForm("email")
	telephone := c.PostForm("telephone")
	password := c.PostForm("password")
	verifyCode := c.PostForm("verifyCode")
	log.Debug("SignupPost", zap.String("email", email), zap.String("telephone", telephone))

	// 验证基本字段
	if len(email) == 0 || len(password) == 0 {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "邮箱或密码不能为空",
			"cfg":       config.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}

	// 验证图片验证码
	s := sessions.Default(c)
	captchaId := s.Get(common.SessionCaptcha)
	if captchaId == nil {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "请先获取验证码",
			"cfg":       config.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}

	if !captcha.VerifyString(captchaId.(string), verifyCode) {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "验证码错误",
			"cfg":       config.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}

	// 验证成功后删除验证码
	s.Delete(common.SessionCaptcha)
	s.Save()

	// 使用bcrypt哈希密码
	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "密码处理失败",
			"cfg":       config.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}

	user := &models.User{
		Email:     email,
		Telephone: telephone,
		Password:  hashedPassword,
		IsAdmin:   true,
	}
	err = user.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "email already exists",
			"cfg":       config.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/signin")
}
