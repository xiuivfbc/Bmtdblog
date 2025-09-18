package models

import "time"

type SmmsFile struct {
	ID        uint       `gorm:"primarykey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	FileName  string     `json:"filename"`
	StoreName string     `json:"storename"`
	Size      int        `json:"size"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Hash      string     `json:"hash"`
	Delete    string     `json:"delete"`
	Url       string     `json:"url"`
	Path      string     `json:"path"`
}

func (sf SmmsFile) Insert() (err error) {
	err = DB.Create(&sf).Error
	return
}
