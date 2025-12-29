package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type PostTag struct {
	ID        uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	PostId    uint       `gorm:"column:post_id;type:uint;uniqueIndex:uk_post_tag" json:"post_id"` // post id
	TagId     uint       `gorm:"column:tag_id;type:uint;uniqueIndex:uk_post_tag" json:"tag_id"`   // tag id
	CreatedAt *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
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
