package router

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/api/content"
	"github.com/xiuivfbc/bmtdblog/internal/api/interaction"
	s "github.com/xiuivfbc/bmtdblog/internal/api/system"
	"github.com/xiuivfbc/bmtdblog/internal/api/upload"
	"github.com/xiuivfbc/bmtdblog/internal/api/user"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func DefineRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 添加全链路追踪中间件（最先执行）
	router.Use(system.TraceMiddleware())

	setTemplate(router)
	setSessions(router)
	router.Use(SharedData())

	router.Static("/static", filepath.Join(common.GetCurrentDirectory(), system.GetConfiguration().PublicDir))

	// favicon.ico路由
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File(filepath.Join(common.GetCurrentDirectory(), system.GetConfiguration().PublicDir, "favicon.ico"))
	})

	router.NoRoute(common.Handle404)
	router.GET("/", content.IndexGet)
	router.GET("/index", content.IndexGet)
	router.GET("/rss", content.RssGet)

	if system.GetConfiguration().SignupEnabled {
		router.GET("/signup", user.SignupGet)
		router.POST("/signup", user.SignupPost)
	}
	// user signin and logout
	router.GET("/signin", user.SigninGet)
	router.POST("/signin", user.SigninPost)
	router.GET("/logout", user.LogoutGet)
	router.GET("/oauth2callback", user.Oauth2Callback)
	router.GET("/auth/:authType", user.AuthGet)
	// captcha
	router.GET("/captcha", user.CaptchaGet)

	visitor := router.Group("/visitor")
	visitor.Use(AuthRequired(false))
	{
		visitor.POST("/new_comment", interaction.CommentPost)
		visitor.POST("/comment/:id/delete", interaction.CommentDelete)
	}

	// subscriber
	router.GET("/subscribe", interaction.SubscribeGet)
	router.POST("/subscribe", interaction.Subscribe)
	router.GET("/active", interaction.ActiveSubscriber) // 暂时没用
	router.GET("/unsubscribe", interaction.UnSubscribe)

	router.GET("/page/:id", content.PageGet)
	router.GET("/post/:id", content.PostGet)
	router.GET("/tag/:tag", content.TagGet)
	router.GET("/archives/:year/:month", content.ArchiveGet)

	// 搜索相关路由
	router.GET("/search", content.SearchGet)
	router.GET("/search/index", content.SearchIndexGet)
	router.GET("/api/search/suggestions", content.SearchSuggestionsAPI)

	router.GET("/link/:id", s.LinkGet)

	authorized := router.Group("/admin")
	authorized.Use(AuthRequired(true))
	{
		// index
		authorized.GET("/index", content.AdminIndex)

		// image upload
		authorized.POST("/upload", upload.Upload)

		// page
		authorized.GET("/page", content.PageIndex)
		authorized.GET("/new_page", content.PageNew)
		authorized.POST("/new_page", content.PageCreate)
		authorized.GET("/page/:id/edit", content.PageEdit)
		authorized.POST("/page/:id/edit", content.PageUpdate)
		authorized.POST("/page/:id/publish", content.PagePublish)
		authorized.POST("/page/:id/delete", content.PageDelete)

		// post
		authorized.GET("/post", content.PostIndex)
		authorized.GET("/new_post", content.PostNew)
		authorized.POST("/new_post", content.PostCreate)
		authorized.GET("/post/:id/edit", content.PostEdit)
		authorized.POST("/post/:id/edit", content.PostUpdate)
		authorized.POST("/post/:id/publish", content.PostPublish)
		authorized.POST("/post/:id/delete", content.PostDelete)

		// tag
		authorized.POST("/new_tag", content.TagCreate)

		//
		authorized.GET("/user", user.UserIndex)
		authorized.POST("/user/:id/lock", user.UserLock)

		// profile
		authorized.GET("/profile", user.ProfileGet)
		authorized.POST("/profile", user.ProfileUpdate)
		authorized.POST("/profile/email/bind", user.BindEmail)
		authorized.POST("/profile/email/unbind", user.UnbindEmail)
		authorized.POST("/profile/github/unbind", user.UnbindGithub)

		// subscriber
		authorized.GET("/subscriber", interaction.SubscriberIndex)
		authorized.POST("/subscriber", interaction.SubscriberPost)
		authorized.POST("/unsubscribe", interaction.UnSubscribe)

		// link
		authorized.GET("/link", s.LinkIndex)
		authorized.POST("/new_link", s.LinkCreate)
		authorized.POST("/link/:id/edit", s.LinkUpdate)
		authorized.POST("/link/:id/delete", s.LinkDelete)

		// comment
		authorized.POST("/comment/:id", interaction.CommentRead)
		authorized.POST("/read_all", interaction.CommentReadAll)

		// backup
		authorized.GET("/backup", s.BackupPost)
		authorized.POST("/restore", s.RestorePost)

		// mail
		authorized.POST("/new_mail", s.SendMail)
		authorized.POST("/new_batchmail", s.SendBatchMail)

		// email queue
		authorized.GET("/email-queue", s.EmailQueueManage)
		authorized.GET("/email-queue/status", s.EmailQueueStatus)
		authorized.POST("/email-queue/retry", s.RetryFailedEmails)
		authorized.POST("/email-queue/clear", s.ClearFailedEmails)
	}
	return router
}

//+++++++++++++ middlewares +++++++++++++++++++++++

func SharedData() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if uID := session.Get(common.SessionKey); uID != nil {
			user, err := models.GetUser(uID)
			if err == nil {
				c.Set(common.ContextUserKey, user)
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
		if user, _ := c.Get(common.ContextUserKey); user != nil {
			if u, ok := user.(*models.User); ok && (!adminScope || u.IsAdmin) {
				c.Next()
				return
			}
		}
		system.LogWarn(c, "User not authorized to visit", "uri", c.Request.RequestURI)
		c.HTML(http.StatusForbidden, "errors/error.html", gin.H{
			"message": "Forbidden!",
		})
		c.Abort()
	}
}

func setTemplate(engine *gin.Engine) {
	funcMap := template.FuncMap{
		"dateFormat": common.DateFormat,
		"substring":  common.Substring,
		"isOdd":      common.IsOdd,
		"isEven":     common.IsEven,
		"truncate":   common.Truncate,
		"length":     common.Len,
		"add":        common.Add,
		"sub":        common.Sub,
		"minus":      common.Minus,
		"multiply":   common.Multiply,
		"seq":        common.Seq,
		"listtag":    common.ListTag,
	}
	engine.SetFuncMap(funcMap)
	engine.LoadHTMLGlob(common.GetCurrentDirectory() + "/front/views/**/*.html")
}

func setSessions(router *gin.Engine) {
	config := system.GetConfiguration()

	var store sessions.Store
	var err error

	// 检查是否启用Redis
	if config.Redis.Enabled {
		system.Logger.Info("使用Redis存储Session", "addr", config.Redis.Addr)

		// 创建Redis存储
		if config.Redis.Password != "" {
			// 有密码时使用密码认证
			store, err = redis.NewStore(
				config.Redis.PoolSize, // 连接池大小
				"tcp",                 // 网络类型
				config.Redis.Addr,     // Redis地址
				config.Redis.Password, // Redis密码
				config.SessionSecret,  // session密钥
			)
		} else {
			// 无密码时不传递密码参数
			store, err = redis.NewStoreWithDB(
				config.Redis.PoolSize,         // 连接池大小
				"tcp",                         // 网络类型
				config.Redis.Addr,             // Redis地址
				"",                            // username(Redis通常不需要)
				"",                            // 无密码
				strconv.Itoa(config.Redis.DB), // 数据库编号
				[]byte(config.SessionSecret),  // session密钥
			)
		}

		if err != nil {
			system.Logger.Error("Redis Session存储初始化失败，回退到Cookie存储", "error", err)
			// 回退到Cookie存储
			store = cookie.NewStore([]byte(config.SessionSecret))
		} else {
			system.Logger.Info("Redis Session存储初始化成功")
		}
	} else {
		system.Logger.Info("使用Cookie存储Session")
		// 使用Cookie存储
		store = cookie.NewStore([]byte(config.SessionSecret))
	}

	// 设置Session选项
	store.Options(sessions.Options{
		HttpOnly: true,
		MaxAge:   7 * 86400, // 7天有效期
		Path:     "/",
		Secure:   false, // 生产环境建议设为true (需要HTTPS)
	})

	router.Use(sessions.Sessions("gin-session", store))
}
