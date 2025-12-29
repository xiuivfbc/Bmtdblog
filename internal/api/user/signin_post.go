package user

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 用户登录
// @Description 使用用户名和密码登录系统
// @Tags 用户认证
// @Accept x-www-form-urlencoded
// @Produce html
// @Param username formData string true "用户名"
// @Param password formData string true "密码"
// @Success 302 {string} string "登录成功，重定向到相应页面"
// @Failure 200 {string} string "登录失败，返回错误信息"
// @Router /signin [post]
func SigninPost(c *gin.Context) {
	var (
		err  error
		user *models.User
	)
	username := c.PostForm("username")
	password := c.PostForm("password")
	log.Debug("SigninPost", zap.String("username", username))
	if username == "" || password == "" {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "username or password cannot be null",
			"cfg":     config.GetConfiguration(),
		})
		return
	}

	// 使用优化的登录查询，利用联合索引
	user, err = models.GetUserForLogin(username)
	if err != nil {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "invalid username or password",
			"cfg":     config.GetConfiguration(),
		})
		return
	}

	// 使用bcrypt验证密码
	if common.CheckPassword(password, user.Password) != nil {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "invalid username or password",
			"cfg":     config.GetConfiguration(),
		})
		return
	}
	if user.LockState {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "Your account have been locked",
			"cfg":     config.GetConfiguration(),
		})
		return
	}
	s := sessions.Default(c)
	s.Clear()
	s.Set(common.SessionKey, user.ID)
	s.Save()
	if user.IsAdmin {
		c.Redirect(http.StatusMovedPermanently, "/admin/index")
	} else {
		c.Redirect(http.StatusMovedPermanently, "/")
	}
}
