package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type SmmsFile struct {
	ID        uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	FileName  string     `gorm:"column:file_name;type:varchar(255)" json:"filename"`
	StoreName string     `gorm:"column:store_name;type:varchar(255)" json:"storename"`
	Size      int        `gorm:"column:size;type:int" json:"size"`
	Width     int        `gorm:"column:width;type:int" json:"width"`
	Height    int        `gorm:"column:height;type:int" json:"height"`
	Hash      string     `gorm:"column:hash;type:varchar(255)" json:"hash"`
	Delete    string     `gorm:"column:delete;type:varchar(255)" json:"delete"`
	Url       string     `gorm:"column:url;type:varchar(500)" json:"url"`
	Path      string     `gorm:"column:path;type:varchar(255)" json:"path"`
	CreatedAt *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (sf SmmsFile) Insert() (err error) {
	DB := dao.GetMysqlDB()
	err = DB.Create(&sf).Error
	return
}
