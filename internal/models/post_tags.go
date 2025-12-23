package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type PostTag struct {
	ID        uint       `gorm:"primarykey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	PostId    uint       `gorm:"uniqueIndex:uk_post_tag"` // post id
	TagId     uint       `gorm:"uniqueIndex:uk_post_tag"` // tag id
}

// post_tags
func (pt *PostTag) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.FirstOrCreate(pt, "post_id = ? and tag_id = ?", pt.PostId, pt.TagId).Error
}

func DeletePostTagByPostId(postId uint) error {
	DB := dao.GetMysqlDB()
	return DB.Delete(&PostTag{}, "post_id = ?", postId).Error
}
