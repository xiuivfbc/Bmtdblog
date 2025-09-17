package models

import "time"

type PostTag struct {
	ID        uint       `gorm:"primarykey"`
	CreatedAt *time.Time `gorm:"default:null"`
	UpdatedAt *time.Time `gorm:"default:null"`
	PostId    uint       `gorm:"uniqueIndex:uk_post_tag"` // post id
	TagId     uint       `gorm:"uniqueIndex:uk_post_tag"` // tag id
}

// post_tags
func (pt *PostTag) Insert() error {
	return DB.FirstOrCreate(pt, "post_id = ? and tag_id = ?", pt.PostId, pt.TagId).Error
}

func DeletePostTagByPostId(postId uint) error {
	return DB.Delete(&PostTag{}, "post_id = ?", postId).Error
}
