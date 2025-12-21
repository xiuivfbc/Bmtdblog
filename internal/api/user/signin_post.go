package user

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func SigninPost(c *gin.Context) {
	var (
		err  error
		user *models.User
	)
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "username or password cannot be null",
			"cfg":     system.GetConfiguration(),
		})
		return
	}

	// 使用优化的登录查询，利用联合索引
	user, err = models.GetUserForLogin(username)
	if err != nil {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "invalid username or password",
			"cfg":     system.GetConfiguration(),
		})
		return
	}

	// 使用bcrypt验证密码
	if common.CheckPassword(password, user.Password) != nil {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "invalid username or password",
			"cfg":     system.GetConfiguration(),
		})
		return
	}
	if user.LockState {
		c.HTML(http.StatusOK, "auth/signin.html", gin.H{
			"message": "Your account have been locked",
			"cfg":     system.GetConfiguration(),
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
