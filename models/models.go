package models

import (
	"github.com/glebarez/sqlite"
	"github.com/wangsongyan/wblog/system"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
		cfg = system.GetConfiguration()
	)
	// 支持 sqlite 和 mysql
	if cfg.Database.Dialect == "sqlite" {
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	} else if cfg.Database.Dialect == "mysql" {
		db, err = gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	}
	if err != nil {
		return nil, err
	}
	DB = db
	//db.LogMode(true)
	db.AutoMigrate(&Page{}, &Post{}, &Tag{}, &PostTag{}, &User{}, &Comment{}, &Subscriber{}, &Link{}, &SmmsFile{})
	return db, err
}
