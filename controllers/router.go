package controllers

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/helpers"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
)

func DefineRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	setTemplate(router)
	setSessions(router)
	router.Use(SharedData())

	router.Static("/static", filepath.Join(helpers.GetCurrentDirectory(), system.GetConfiguration().PublicDir))

	router.NoRoute(Handle404)
	router.GET("/", IndexGet)
	router.GET("/index", IndexGet)
	router.GET("/rss", RssGet)

	if system.GetConfiguration().SignupEnabled {
		router.GET("/signup", SignupGet)
		router.POST("/signup", SignupPost)
	}
	// user signin and logout
	router.GET("/signin", SigninGet)
	router.POST("/signin", SigninPost)
	router.GET("/logout", LogoutGet)
	router.GET("/oauth2callback", Oauth2Callback)
	router.GET("/auth/:authType", AuthGet)

	// captcha
	router.GET("/captcha", CaptchaGet)

	visitor := router.Group("/visitor")
	visitor.Use(AuthRequired(false))
	{
		visitor.POST("/new_comment", CommentPost)
		visitor.POST("/comment/:id/delete", CommentDelete)
	}

	// subscriber
	router.GET("/subscribe", SubscribeGet)
	router.POST("/subscribe", Subscribe)
	router.GET("/active", ActiveSubscriber) // 暂时没用
	router.GET("/unsubscribe", UnSubscribe)

	router.GET("/page/:id", PageGet)
	router.GET("/post/:id", PostGet)
	router.GET("/tag/:tag", TagGet)
	router.GET("/archives/:year/:month", ArchiveGet)

	router.GET("/link/:id", LinkGet)

	authorized := router.Group("/admin")
	authorized.Use(AuthRequired(true))
	{
		// index
		authorized.GET("/index", AdminIndex)

		// image upload
		authorized.POST("/upload", Upload)

		// page
		authorized.GET("/page", PageIndex)
		authorized.GET("/new_page", PageNew)
		authorized.POST("/new_page", PageCreate)
		authorized.GET("/page/:id/edit", PageEdit)
		authorized.POST("/page/:id/edit", PageUpdate)
		authorized.POST("/page/:id/publish", PagePublish)
		authorized.POST("/page/:id/delete", PageDelete)

		// post
		authorized.GET("/post", PostIndex)
		authorized.GET("/new_post", PostNew)
		authorized.POST("/new_post", PostCreate)
		authorized.GET("/post/:id/edit", PostEdit)
		authorized.POST("/post/:id/edit", PostUpdate)
		authorized.POST("/post/:id/publish", PostPublish)
		authorized.POST("/post/:id/delete", PostDelete)

		// tag
		authorized.POST("/new_tag", TagCreate)

		//
		authorized.GET("/user", UserIndex)
		authorized.POST("/user/:id/lock", UserLock)

		// profile
		authorized.GET("/profile", ProfileGet)
		authorized.POST("/profile", ProfileUpdate)
		authorized.POST("/profile/email/bind", BindEmail)
		authorized.POST("/profile/email/unbind", UnbindEmail)
		authorized.POST("/profile/github/unbind", UnbindGithub)

		// subscriber
		authorized.GET("/subscriber", SubscriberIndex)
		authorized.POST("/subscriber", SubscriberPost)
		authorized.POST("/unsubscribe", UnSubscribe)

		// link
		authorized.GET("/link", LinkIndex)
		authorized.POST("/new_link", LinkCreate)
		authorized.POST("/link/:id/edit", LinkUpdate)
		authorized.POST("/link/:id/delete", LinkDelete)

		// comment
		authorized.POST("/comment/:id", CommentRead)
		authorized.POST("/read_all", CommentReadAll)

		// backup
		authorized.GET("/backup", BackupPost)
		authorized.POST("/restore", RestorePost)

		// mail
		authorized.POST("/new_mail", SendMail)
		authorized.POST("/new_batchmail", SendBatchMail)
	}
	return router
}

//+++++++++++++ middlewares +++++++++++++++++++++++

func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uID := session.Get(SessionKey); uID != nil {
			user, err := models.GetUser(uID)
			if err == nil {
				c.Set(ContextUserKey, user)
			}
		}
		if system.GetConfiguration().SignupEnabled {
			c.Set("SignupEnabled", true)
		}
		c.Next()
	}
}

func AuthRequired(adminScope bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, _ := c.Get(ContextUserKey); user != nil {
			if u, ok := user.(*models.User); ok && (!adminScope || u.IsAdmin) {
				c.Next()
				return
			}
		}
		slog.Warn("User not authorized to visit", "uri", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}

func setTemplate(engine *gin.Engine) {
	funcMap := template.FuncMap{
		"dateFormat": helpers.DateFormat,
		"substring":  helpers.Substring,
		"isOdd":      helpers.IsOdd,
		"isEven":     helpers.IsEven,
		"truncate":   helpers.Truncate,
		"length":     helpers.Len,
		"add":        helpers.Add,
		"minus":      helpers.Minus,
		"listtag":    helpers.ListTag,
	}
	engine.SetFuncMap(funcMap)
	engine.LoadHTMLGlob(helpers.GetCurrentDirectory() + "/views/**/*.html")
}

func setSessions(router *gin.Engine) {
	config := system.GetConfiguration()
	store := cookie.NewStore([]byte(config.SessionSecret))
	store.Options(sessions.Options{HttpOnly: true, MaxAge: 7 * 86400, Path: "/"})
	router.Use(sessions.Sessions("gin-session", store))
}
