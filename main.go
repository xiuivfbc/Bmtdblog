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
	"github.com/wangsongyan/wblog/controllers"
	"github.com/wangsongyan/wblog/models"
	"github.com/wangsongyan/wblog/system"
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
	initSomething()

	router = controllers.DefineRouter()

	//Periodic tasks
	gocron.Every(1).Day().Do(controllers.CreateXMLSitemap)
	gocron.Every(7).Days().Do(controllers.Backup)
	gocron.Start()

	fmt.Println("Welcome to bmtdblog!")
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go clean()
	srv = &http.Server{
		Addr:    system.GetConfiguration().Addr,
		Handler: router,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "err", err)
	}

	defer func() {
		f.Close()
		database, _ := db.DB()
		database.Close()
	}()
}

// canclue the program gracefully
func clean() {
	<-c
	fmt.Println("Cleaning...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "err", err)
	}
	gocron.Clear()
	dbInstance, _ := db.DB()
	_ = dbInstance.Close()
	f.Close()
	cancel()
	os.Exit(0)
}

// init configuration , logger and database
func initSomething() {
	// logger
	logDir := "log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}
	logFile := filepath.Join(logDir, fmt.Sprintf("wblog-%s.log", time.Now().Format("20060102-150405")))
	f, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(f, nil))
	slog.SetDefault(logger)

	//configuration
	configFilePath := flag.String("C", "conf/conf.toml", "config file path")
	flag.Parse()
	if err := system.LoadConfiguration(*configFilePath); err != nil {
		logger.Error("err parsing config log file", "err", err)
		f.Close()
		os.Exit(1)
	}

	//database
	db, err = models.InitDB()
	if err != nil {
		logger.Error("err open databases", "err", err)
		f.Close()
		os.Exit(1)
	}
}
