package router

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/xiuivfbc/bmtdblog/docs" // 导入生成的docs包
	"github.com/xiuivfbc/bmtdblog/internal/api/backup"
	"github.com/xiuivfbc/bmtdblog/internal/api/comment"
	"github.com/xiuivfbc/bmtdblog/internal/api/content"
	"github.com/xiuivfbc/bmtdblog/internal/api/email"
	"github.com/xiuivfbc/bmtdblog/internal/api/link"
	"github.com/xiuivfbc/bmtdblog/internal/api/queue"
	"github.com/xiuivfbc/bmtdblog/internal/api/subscribe"
	"github.com/xiuivfbc/bmtdblog/internal/api/upload"
	"github.com/xiuivfbc/bmtdblog/internal/api/user"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/middleware" // 导入新的中间件包
)

func DefineRouter() *gin.Engine {
	// 初始化Gin引擎
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// ------------------------------
	// 中间件配置
	// ------------------------------
	router.Use(middleware.TraceMiddleware()) // 全链路追踪中间件（最先执行）
	setSessions(router)                      // 会话配置
	router.Use(middleware.SharedData())      // 共享数据中间件

	// ------------------------------
	// 模板配置
	// ------------------------------
	setTemplate(router)

	// ------------------------------
	// 静态资源处理
	// ------------------------------
	router.Static("/static", filepath.Join(common.GetCurrentDirectory(), config.GetConfiguration().PublicDir))
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File(filepath.Join(common.GetCurrentDirectory(), config.GetConfiguration().PublicDir, "favicon.ico"))
	})

	// ------------------------------
	// 基础路由
	// ------------------------------
	router.NoRoute(common.Handle404) // 404处理
	router.GET("/", content.IndexGet)
	router.GET("/index", content.IndexGet)
	router.GET("/rss", content.RssGet)

	// ------------------------------
	// API文档路由
	// ------------------------------
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ------------------------------
	// 用户认证路由
	// ------------------------------
	if config.GetConfiguration().SignupEnabled {
		router.GET("/signup", user.SignupGet)
		router.POST("/signup", user.SignupPost)
	}
	router.GET("/signin", user.SigninGet)
	router.POST("/signin", user.SigninPost)
	router.GET("/logout", user.LogoutGet)
	router.GET("/oauth2callback", user.Oauth2Callback)
	router.GET("/auth/:authType", user.AuthGet)
	router.GET("/captcha", user.CaptchaGet) // 验证码

	// ------------------------------
	// 内容浏览路由
	// ------------------------------
	router.GET("/page/:id", content.PageGet)                 // 单页
	router.GET("/post/:id", content.PostGet)                 // 文章详情
	router.GET("/tag/:tag", content.TagGet)                  // 标签页
	router.GET("/archives/:year/:month", content.ArchiveGet) // 归档页
	router.GET("/link/:id", link.LinkGet)                    // 友情链接详情

	// ------------------------------
	// 搜索功能路由
	// ------------------------------
	router.GET("/search", content.SearchGet)
	router.GET("/search/index", content.SearchIndexGet)
	// 搜索建议API（旧版本，建议使用新版本）
	router.GET("/api/search/suggestions", content.SearchSuggestionsAPI)
	// 搜索建议API（新版本，添加版本控制）
	router.GET("/api/v1/search/suggestions", content.SearchSuggestionsAPI)

	// ------------------------------
	// 交互功能路由
	// ------------------------------
	// 评论相关（需要用户认证但不需要管理员权限）
	visitor := router.Group("/visitor")
	visitor.Use(middleware.AuthRequired(false))
	{
		visitor.POST("/new_comment", comment.CommentPost)          // 发表评论
		visitor.POST("/comment/:id/delete", comment.CommentDelete) // 删除自己的评论
	}

	// 订阅相关
	router.GET("/subscribe", subscribe.SubscribeGet)  // 订阅页面
	router.POST("/subscribe", subscribe.Subscribe)    // 提交订阅
	router.GET("/active", subscribe.ActiveSubscriber) // 激活订阅者（暂未使用）
	router.GET("/unsubscribe", subscribe.UnSubscribe) // 取消订阅

	// ------------------------------
	// 管理后台路由（需要管理员权限）
	// ------------------------------
	admin := router.Group("/admin")
	admin.Use(middleware.AuthRequired(true))
	{
		// 管理后台首页
		admin.GET("/index", content.AdminIndex)

		// 内容管理
		admin.GET("/page", content.PageIndex)                // 页面列表
		admin.GET("/new_page", content.PageNew)              // 创建页面
		admin.POST("/new_page", content.PageCreate)          // 提交页面
		admin.GET("/page/:id/edit", content.PageEdit)        // 编辑页面
		admin.POST("/page/:id/edit", content.PageUpdate)     // 更新页面
		admin.POST("/page/:id/publish", content.PagePublish) // 发布页面
		admin.POST("/page/:id/delete", content.PageDelete)   // 删除页面

		admin.GET("/post", content.PostIndex)                // 文章列表
		admin.GET("/new_post", content.PostNew)              // 创建文章
		admin.POST("/new_post", content.PostCreate)          // 提交文章
		admin.GET("/post/:id/edit", content.PostEdit)        // 编辑文章
		admin.POST("/post/:id/edit", content.PostUpdate)     // 更新文章
		admin.POST("/post/:id/publish", content.PostPublish) // 发布文章
		admin.POST("/post/:id/delete", content.PostDelete)   // 删除文章

		admin.POST("/new_tag", content.TagCreate) // 创建标签

		// 用户管理
		admin.GET("/user", user.UserIndex)          // 用户列表
		admin.POST("/user/:id/lock", user.UserLock) // 锁定/解锁用户

		// 个人资料管理
		admin.GET("/profile", user.ProfileGet)                  // 个人资料页面
		admin.POST("/profile", user.ProfileUpdate)              // 更新个人资料
		admin.POST("/profile/email/bind", user.BindEmail)       // 绑定邮箱
		admin.POST("/profile/email/unbind", user.UnbindEmail)   // 解绑邮箱
		admin.POST("/profile/github/unbind", user.UnbindGithub) // 解绑Github

		// 订阅者管理
		admin.GET("/subscriber", subscribe.SubscriberIndex) // 订阅者列表
		admin.POST("/subscriber", subscribe.SubscriberPost) // 提交订阅者
		admin.POST("/unsubscribe", subscribe.UnSubscribe)   // 取消订阅

		// 友情链接管理
		admin.GET("/link", link.LinkIndex)              // 友情链接列表
		admin.POST("/new_link", link.LinkCreate)        // 创建友情链接
		admin.POST("/link/:id/edit", link.LinkUpdate)   // 更新友情链接
		admin.POST("/link/:id/delete", link.LinkDelete) // 删除友情链接

		// 评论管理
		admin.POST("/comment/:id", comment.CommentRead) // 标记评论为已读
		admin.POST("/read_all", comment.CommentReadAll) // 标记所有评论为已读

		// 备份与恢复
		admin.GET("/backup", backup.BackupPost)    // 备份数据
		admin.POST("/restore", backup.RestorePost) // 恢复数据

		// 邮件管理
		admin.POST("/new_mail", email.SendMail)           // 发送单封邮件
		admin.POST("/new_batchmail", email.SendBatchMail) // 发送批量邮件

		// 邮件队列管理
		admin.GET("/email-queue", queue.EmailQueueManage)         // 邮件队列管理页面
		admin.GET("/email-queue/status", queue.EmailQueueStatus)  // 邮件队列状态
		admin.POST("/email-queue/retry", queue.RetryFailedEmails) // 重试失败邮件
		admin.POST("/email-queue/clear", queue.ClearFailedEmails) // 清除失败邮件

		// 上传管理
		admin.POST("/upload", upload.Upload) // 上传文件
	}

	return router
}
