package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/claudiu/gocron"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/controllers"
	"github.com/xiuivfbc/bmtdblog/helpers"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
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
	initializeApplication()
	startApplication()
}

// startApplication 启动应用程序服务
func startApplication() {
	router = controllers.DefineRouter()

	// 启动定时任务
	setupPeriodicTasks()

	fmt.Println("Welcome to bmtdblog!")

	// 设置优雅关闭
	setupGracefulShutdown()

	// 创建并启动服务器
	cfg := system.GetConfiguration()
	srv = &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	// 启动服务器
	if err := system.StartServer(srv); err != nil && err != http.ErrServerClosed {
		system.Logger.Error("Server error", "err", err)
	}

	// 清理资源
	defer cleanupResources()
}

// setupPeriodicTasks 设置定时任务
func setupPeriodicTasks() {
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Every(7).Days().Do(controllers.Backup)
	gocron.Start()
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go clean()
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

// canclue the program gracefully
func clean() {
	<-c
	fmt.Println("Cleaning...")

	// 停止ES任务队列
	models.StopESTaskQueue()

	// 停止邮件队列
	system.StopEmailQueue()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		system.Logger.Error("HTTP server shutdown error", "err", err)
	}
	gocron.Clear()
	dbInstance, _ := db.DB()
	_ = dbInstance.Close()
	f.Close()
	cancel()
	os.Exit(0)
}

// initializeApplication 初始化应用程序
func initializeApplication() {
	// system.logger
	logDir := "slog"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}
	logFile := filepath.Join(logDir, fmt.Sprintf("Bmtdblog-%s.log", time.Now().Format("20060102-150405")))
	f, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	// 调试用设置
	var opts *slog.HandlerOptions = nil
	opts = &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	system.Logger = slog.New(slog.NewJSONHandler(f, opts))
	slog.SetDefault(system.Logger)

	//configuration
	configFilePath := flag.String("C", "conf/conf_mine.toml", "config file path")
	flag.Parse()
	if err := system.LoadConfiguration(*configFilePath); err != nil {
		system.Logger.Error("err parsing config log file", "err", err)
		f.Close()
		os.Exit(1)
	}

	//database
	db, err = models.InitDB()
	if err != nil {
		system.Logger.Error("err open databases", "err", err)
		f.Close()
		os.Exit(1)
	}

	// Redis缓存初始化
	if err := system.InitRedis(); err != nil {
		system.Logger.Error("Redis initialization failed", "err", err)
		// Redis失败不退出程序，允许降级运行
	}

	// ElasticSearch搜索初始化
	if system.GetConfiguration().Elasticsearch.Enabled {
		if err := system.InitElasticsearch(); err != nil {
			system.Logger.Error("ElasticSearch initialization failed", "err", err)
			// ES失败不退出程序，允许降级运行
		} else {
			// 启动ES批量任务队列
			models.InitESTaskQueue()
		}
	} else {
		system.Logger.Info("ElasticSearch功能已禁用")
	}

	// 邮件队列初始化
	workerCount := 3 // 启动3个邮件工作者
	if err := system.InitEmailQueue(workerCount); err != nil {
		system.Logger.Error("EmailQueue initialization failed", "err", err)
		// 邮件队列失败不退出程序，允许降级运行
	} else {
		// 设置邮件发送回调函数
		system.SetEmailSender(helpers.SendMail)
	}
}
