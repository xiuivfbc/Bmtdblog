package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Link struct {
	ID        uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	Name      string     `gorm:"column:name;type:varchar(100)" json:"name"`
	Url       string     `gorm:"column:url;type:varchar(255)" json:"url"`
	Sort      int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	View      int        `gorm:"column:view;type:int;default:0" json:"view"`
	CreatedAt *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (link *Link) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.FirstOrCreate(link, "url = ?", link.Url).Error
}

func (link *Link) Update() error {
	DB := dao.GetMysqlDB()
	return DB.Model(link).Updates(map[string]any{
		"Name": link.Name,
		"Url":  link.Url,
		"Sort": link.Sort,
	}).Error
}

func (link *Link) Delete() error {
	DB := dao.GetMysqlDB()
	return DB.Delete(link).Error
}

func ListLinks() ([]*Link, error) {
	var links []*Link
	DB := dao.GetMysqlDB()
	err := DB.Order("sort asc").Find(&links).Error
	return links, err
}

func MustListLinks() []*Link {
	links, _ := ListLinks()
	return links
}

func GetLinkById(id uint) (*Link, error) {
	var link Link
	DB := dao.GetMysqlDB()
	err := DB.FirstOrCreate(&link, "id = ?", id).Error
	return &link, err
}
