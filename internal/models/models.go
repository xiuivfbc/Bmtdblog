package models

import (
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
		cfg = config.GetConfiguration()
	)
	db, err = gorm.Open(mysql.Open(cfg.Database.Dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	DB = db
	db.AutoMigrate(
		&Page{},
		&Post{},
		&Tag{},
		&PostTag{},
		&User{},
		&Comment{},
		&Subscriber{},
		&Link{},
		&SmmsFile{},
		&ESSyncStatus{},
	)
	return db, err
}
