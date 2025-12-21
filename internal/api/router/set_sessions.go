package router

import (
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

func setSessions(router *gin.Engine) {
	cfg := config.GetConfiguration()

	var store sessions.Store
	var err error

	// 检查是否启用Redis
	if cfg.Redis.Enabled {
		config.Logger.Info("使用Redis存储Session", "addr", cfg.Redis.Addr)

		// 创建Redis存储
		if cfg.Redis.Password != "" {
			// 有密码时使用密码认证
			store, err = redis.NewStore(
				cfg.Redis.PoolSize, // 连接池大小
				"tcp",              // 网络类型
				cfg.Redis.Addr,     // Redis地址
				cfg.Redis.Password, // Redis密码
				cfg.SessionSecret,  // session密钥
			)
		} else {
			// 无密码时不传递密码参数
			store, err = redis.NewStoreWithDB(
				cfg.Redis.PoolSize,         // 连接池大小
				"tcp",                      // 网络类型
				cfg.Redis.Addr,             // Redis地址
				"",                         // username(Redis通常不需要)
				"",                         // 无密码
				strconv.Itoa(cfg.Redis.DB), // 数据库编号
				[]byte(cfg.SessionSecret),  // session密钥
			)
		}

		if err != nil {
			config.Logger.Error("Redis Session存储初始化失败，回退到Cookie存储", "error", err)
			// 回退到Cookie存储
			store = cookie.NewStore([]byte(cfg.SessionSecret))
		} else {
			config.Logger.Info("Redis Session存储初始化成功")
		}
	} else {
		config.Logger.Info("使用Cookie存储Session")
		// 使用Cookie存储
		store = cookie.NewStore([]byte(cfg.SessionSecret))
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
