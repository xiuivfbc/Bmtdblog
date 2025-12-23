package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Tag struct {
	ID        uint       `gorm:"primarykey"`
	CreatedAt *time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	Name      string
	Total     int `gorm:"->"`
}

func (tag *Tag) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.FirstOrCreate(tag, "name = ?", tag.Name).Error
}

func ListTag() ([]*Tag, error) {
	var tags []*Tag
	DB := dao.GetMysqlDB()
	rows, err := DB.Raw("select t.*,count(*) total from tags t inner join post_tags pt on t.id = pt.tag_id inner join posts p on pt.post_id = p.id where p.is_published = ? group by pt.tag_id", true).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

func MustListTag() []*Tag {
	tags, _ := ListTag()
	return tags
}

func ListTagByPostId(id uint) ([]*Tag, error) {
	var tags []*Tag
	DB := dao.GetMysqlDB()
	rows, err := DB.Raw("select t.* from tags t inner join post_tags pt on t.id = pt.tag_id where pt.post_id = ?", id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var tag Tag
		DB.ScanRows(rows, &tag)
		tags = append(tags, &tag)
	}
	return tags, nil
}

func CountTag() int64 {
	var count int64
	DB := dao.GetMysqlDB()
	DB.Model(&Tag{}).Count(&count)
	return count
}

/*func ListAllTag() ([]*Tag, error) {
	var tags []*Tag
	err := DB.Model(&Tag{}).Find(&tags).Error
	return tags, err
}*/
