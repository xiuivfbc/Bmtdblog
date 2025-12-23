package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Page struct {
	ID          uint       `gorm:"primarykey"`
	CreatedAt   *time.Time `gorm:"autoCreateTime"`
	UpdatedAt   *time.Time `gorm:"autoUpdateTime"`
	Title       string     `gorm:"type:text"`     // title
	Body        string     `gorm:"type:longtext"` // body
	View        int        // view count
	IsPublished bool       // published or not
}

func (page *Page) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.Create(page).Error
}

func (page *Page) Update() error {
	DB := dao.GetMysqlDB()
	return DB.Model(page).Updates(map[string]any{
		"title":        page.Title,
		"body":         page.Body,
		"is_published": page.IsPublished,
	}).Error
}

func (page *Page) UpdateView() error {
	DB := dao.GetMysqlDB()
	return DB.Model(page).Updates(map[string]any{
		"view": page.View,
	}).Error
}

func (page *Page) Delete() error {
	DB := dao.GetMysqlDB()
	return DB.Delete(page).Error
}

func GetPageById(id uint) (*Page, error) {
	var page Page
	DB := dao.GetMysqlDB()
	err := DB.First(&page, "id = ?", id).Error
	return &page, err
}

func ListPublishedPage() ([]*Page, error) {
	return _listPage(true)
}

func ListAllPage() ([]*Page, error) {
	return _listPage(false)
}

func _listPage(published bool) ([]*Page, error) {
	DB := dao.GetMysqlDB()
	var pages []*Page
	var err error
	if published {
		err = DB.Where("is_published = ?", true).Find(&pages).Error
	} else {
		err = DB.Find(&pages).Error
	}
	return pages, err
}

func CountPage() int64 {
	var count int64
	DB := dao.GetMysqlDB()
	DB.Model(&Page{}).Count(&count)
	return count
}
