package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Link struct {
	ID        uint       `gorm:"primarykey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
	Name      string
	Url       string
	Sort      int `gorm:"default:0"`
	View      int
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
