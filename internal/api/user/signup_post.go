package user

import (
	"net/http"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func SignupPost(c *gin.Context) {
	var (
		err error
	)
	email := c.PostForm("email")
	telephone := c.PostForm("telephone")
	password := c.PostForm("password")
	verifyCode := c.PostForm("verifyCode")

	// 验证基本字段
	if len(email) == 0 || len(password) == 0 {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "邮箱或密码不能为空",
			"cfg":       system.GetConfiguration(),
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
			"cfg":       system.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}

	if !captcha.VerifyString(captchaId.(string), verifyCode) {
		c.HTML(http.StatusOK, "auth/signup.html", gin.H{
			"message":   "验证码错误",
			"cfg":       system.GetConfiguration(),
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
			"cfg":       system.GetConfiguration(),
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
			"cfg":       system.GetConfiguration(),
			"email":     email,
			"telephone": telephone,
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/signin")
}
