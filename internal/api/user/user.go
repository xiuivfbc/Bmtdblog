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

type GithubUserInfo struct {
	AvatarURL         string      `json:"avatar_url"`
	Bio               interface{} `json:"bio"`
	Blog              string      `json:"blog"`
	Company           interface{} `json:"company"`
	CreatedAt         string      `json:"created_at"`
	Email             interface{} `json:"email"`
	EventsURL         string      `json:"events_url"`
	Followers         int         `json:"followers"`
	FollowersURL      string      `json:"followers_url"`
	Following         int         `json:"following"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	GravatarID        string      `json:"gravatar_id"`
	Hireable          interface{} `json:"hireable"`
	HTMLURL           string      `json:"html_url"`
	ID                int         `json:"id"`
	Location          interface{} `json:"location"`
	Login             string      `json:"login"`
	Name              interface{} `json:"name"`
	OrganizationsURL  string      `json:"organizations_url"`
	PublicGists       int         `json:"public_gists"`
	PublicRepos       int         `json:"public_repos"`
	ReceivedEventsURL string      `json:"received_events_url"`
	ReposURL          string      `json:"repos_url"`
	SiteAdmin         bool        `json:"site_admin"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	Type              string      `json:"type"`
	UpdatedAt         string      `json:"updated_at"`
	URL               string      `json:"url"`
}

func SigninGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signin.html", gin.H{
		"cfg": system.GetConfiguration(),
	})
}

func SignupGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signup.html", gin.H{
		"cfg": system.GetConfiguration(),
	})
}

func LogoutGet(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Save()
	c.Redirect(http.StatusSeeOther, "/")
}

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

func ProfileGet(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/profile.html", gin.H{
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      system.GetConfiguration(),
	})
}

func ProfileUpdate(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	avatarUrl := c.PostForm("avatarUrl")
	nickName := c.PostForm("nickName")
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	err = user.UpdateProfile(avatarUrl, nickName)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["user"] = models.User{AvatarUrl: avatarUrl, NickName: nickName}
}

func BindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	email := c.PostForm("email")
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	if len(user.Email) > 0 {
		res["message"] = "email have bound"
		return
	}
	_, err = models.GetUserByUsername(email)
	if err == nil {
		res["message"] = "email have be registered"
		return
	}
	err = user.UpdateEmail(email)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UnbindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	if user.Email == "" {
		res["message"] = "email haven't bound"
		return
	}
	err = user.UpdateEmail("")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UnbindGithub(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	if user.GithubLoginId == "" {
		res["message"] = "github haven't bound"
		return
	}
	user.GithubLoginId = ""
	err = user.UpdateGithubUserInfo()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func UserIndex(c *gin.Context) {
	users, _ := models.ListUsers()
	c.HTML(http.StatusOK, "admin/user.html", gin.H{
		"users":    users,
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      system.GetConfiguration(),
	})
}

func UserLock(c *gin.Context) {
	var (
		err  error
		id   uint
		res  = gin.H{}
		user *models.User
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	user, err = models.GetUser(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	user.LockState = !user.LockState
	err = user.Lock()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
