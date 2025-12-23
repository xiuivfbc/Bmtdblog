package models

import (
	"gorm.io/gorm"
)

// RegisterModels 注册所有模型
func RegisterModels(db *gorm.DB) error {
	models := []interface{}{
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
	}

	// 自动迁移模式
	return db.AutoMigrate(models...)
}
