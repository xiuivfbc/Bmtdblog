// @title Bmtdblog API
// @version 1.0
// @description 一个基于Go和Gin的博客系统API文档
// @termsOfService http://localhost:8090

// @contact.name API Support
// @contact.url http://localhost:8090
// @contact.email 3138910969@qq.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8090
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/claudiu/gocron"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/api/backup"
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	r "github.com/xiuivfbc/bmtdblog/internal/router"
	"github.com/xiuivfbc/bmtdblog/internal/server"
	"gorm.io/gorm"
)

var (
	c      = make(chan os.Signal, 1)
	err    error
	db     *gorm.DB
	f      *os.File
	router *gin.Engine
	srv    *http.Server
)

func main() {
	//configuration
	configFilePath := flag.String("C", "configs/conf_mine.toml", "config file path")
	flag.Parse()
	if err := config.LoadConfiguration(*configFilePath); err != nil {
		fmt.Printf("err parsing config log file: %v\n", err)
		if f != nil {
			f.Close()
		}
		os.Exit(1)
	}

	// 设置日志
	if err := log.Init(); err != nil {
		panic(err)
	}

	//database
	db, err = models.InitDB()
	if err != nil {
		log.Error("err open databases", "err", err)
		if f != nil {
			f.Close()
		}
		os.Exit(1)
	}

	// Redis缓存初始化
	if err := dao.InitRedis(); err != nil {
		log.Error("Redis initialization failed", "err", err)
		// Redis失败不退出程序，允许降级运行
	}

	// ElasticSearch搜索初始化
	if config.GetConfiguration().Elasticsearch.Enabled {
		if err := dao.InitElasticsearch(); err != nil {
			log.Error("ElasticSearch initialization failed", "err", err)
			// ES失败不退出程序，允许降级运行
		} else {
			// 启动ES批量任务队列
			models.InitESTaskQueue()
		}
	} else {
		log.Info("ElasticSearch功能已禁用")
	}

	// 邮件队列初始化
	workerCount := 3 // 启动3个邮件工作者
	if err := dao.InitEmailQueue(workerCount); err != nil {
		log.Error("EmailQueue initialization failed", "err", err)
		// 邮件队列失败不退出程序，允许降级运行
	} else {
		// 设置邮件发送回调函数
		dao.SetEmailSender(common.SendMail)
	}

	router = r.DefineRouter()

	// 启动定时任务
	setupPeriodicTasks()

	fmt.Println("Welcome to bmtdblog!")

	// 设置优雅关闭
	setupGracefulShutdown()

	// 创建并启动服务器
	cfg := config.GetConfiguration()
	srv = &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	// 启动服务器
	if err := server.StartServer(srv); err != nil && err != http.ErrServerClosed {
		log.Error("Server error", "err", err)
	}

	// 清理资源
	defer cleanupResources()
}

// setupPeriodicTasks 设置定时任务
func setupPeriodicTasks() {
	gocron.Every(1).Day().Do(common.CreateXMLSitemap)
	gocron.Every(7).Days().Do(backup.Backup)
	gocron.Start()
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		fmt.Println("Cleaning...")

		// 停止ES任务队列
		models.StopESTaskQueue()

		// 停止邮件队列
		dao.StopEmailQueue()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("HTTP server shutdown error", "err", err)
		}
		gocron.Clear()
		dbInstance, _ := db.DB()
		_ = dbInstance.Close()
		f.Close()
		cancel()
		os.Exit(0)
	}()
}

// cleanupResources 清理资源
func cleanupResources() {
	if f != nil {
		f.Close()
	}
	if db != nil {
		if database, err := db.DB(); err == nil {
			database.Close()
		}
	}
}
