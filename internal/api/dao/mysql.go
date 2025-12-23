package dao

import (
	"fmt"
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"context"
)

var mysqlDB *gorm.DB

type MysqlInitOption struct {
	MaxRetry      int
	RetryInterval time.Duration
	Timeout       time.Duration
}

func InitMysql(conf config.Mysql) {
	opt := MysqlInitOption{
		MaxRetry:      6,
		RetryInterval: 5 * time.Second,
		Timeout:       60 * time.Second,
	}

	err := InitMysqlWithOption(conf, opt)
	if err != nil {
		//log.Errorf("InitMysql failed after %d retries: %v", opt.MaxRetry, err)
		//return
		panic(fmt.Sprintf("InitMysql failed after %d retries: %v", opt.MaxRetry, err))
	}
}

func InitMysqlWithOption(conf config.Mysql, opt MysqlInitOption) error {
	start := time.Now()

	for attempt := 1; attempt <= opt.MaxRetry; attempt++ {

		err := func() (err error) {
			// 捕获 panic，避免程序崩溃
			defer func() {
				if r := recover(); r != nil {
					//log.Errorf("InitMysql panic recovered: %v", r)
					err = fmt.Errorf("panic: %v", r)
				}
			}()

			// 1. 使用配置中的密码
			// TODO: 可以考虑支持从环境变量或安全存储中获取密码
			password := conf.Password

			// 2. 构建连接串
			dsn := fmt.Sprintf(
				"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=UTC",
				conf.User, password, conf.Host, conf.Port, conf.DbName,
			)

			// 3. 连接 MySQL
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
				Logger: logger.Default.LogMode(logger.LogLevel(conf.LogLevel)),
			})
			if err != nil {
				return fmt.Errorf("gorm open error: %w", err)
			}

			// 4. 获取底层 *sql.DB
			sqlDB, err := db.DB()
			if err != nil {
				return fmt.Errorf("sql.DB object error: %w", err)
			}

			// 设置连接池参数
			sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
			sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
			sqlDB.SetConnMaxIdleTime(time.Duration(conf.MaxIdleTime) * time.Minute)

			// 5. Ping 检查服务是否 ready
			pingCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := sqlDB.PingContext(pingCtx); err != nil {
				return fmt.Errorf("mysql ping failed: %w", err)
			}

			// 一切正常
			mysqlDB = db
			return nil
		}()

		if err == nil {
			//log.Infof("InitMysql success (attempt %d)", attempt)
			return nil
		}

		//log.Errorf("InitMysql failed (attempt %d/%d): %v", attempt, opt.MaxRetry, err)

		// 判断是否超时
		if time.Since(start) > opt.Timeout {
			return fmt.Errorf("InitMysql timeout after %v: %w", opt.Timeout, err)
		}

		// 重试等待
		if attempt < opt.MaxRetry {
			//log.Infof("Retry after %v ...", opt.RetryInterval)
			time.Sleep(opt.RetryInterval)
		}
	}

	return fmt.Errorf("InitMysql failed after %d retries", opt.MaxRetry)
}

func GetMysqlDB() (db *gorm.DB) {
	return mysqlDB
}
