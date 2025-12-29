package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Comment struct {
	ID        uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	UserID    uint       `gorm:"column:user_id;type:uint" json:"user_id"`
	Content   string     `gorm:"column:content;type:text" json:"content"`
	PostID    uint       `gorm:"column:post_id;type:uint" json:"post_id"`
	ReadState bool       `gorm:"column:read_state;type:bool;default:false" json:"read_state"`
	NickName  string     `gorm:"-" json:"nick_name"`
	AvatarUrl string     `gorm:"-" json:"avatar_url"`
	GithubUrl string     `gorm:"-" json:"github_url"`
	CreatedAt *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (comment *Comment) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.Create(comment).Error
}

func (comment *Comment) Update() error {
	DB := dao.GetMysqlDB()
	return DB.Model(comment).UpdateColumn("read_state", true).Error
}

func SetAllCommentRead() error {
	DB := dao.GetMysqlDB()
	return DB.Model(&Comment{}).Where("read_state = ?", false).Update("read_state", true).Error
}

func ListUnreadComment() ([]*Comment, error) {
	DB := dao.GetMysqlDB()
	var comments []*Comment
	err := DB.Where("read_state = ?", false).Order("created_at desc").Find(&comments).Error
	return comments, err
}

func MustListUnreadComment() []*Comment {
	comments, _ := ListUnreadComment()
	return comments
}

func (comment *Comment) Delete() error {
	DB := dao.GetMysqlDB()
	return DB.Delete(comment, "user_id = ?", comment.UserID).Error
}

func ListCommentByPostID(id uint) ([]*Comment, error) {
	var comments []*Comment
	DB := dao.GetMysqlDB()
	rows, err := DB.Raw("select c.*,u.github_login_id nick_name,u.avatar_url,u.github_url from comments c inner join users u on c.user_id = u.id where c.post_id = ? order by created_at desc", id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var comment Comment
		DB.ScanRows(rows, &comment)
		comments = append(comments, &comment)
	}
	return comments, err
}

func CountComment() int64 {
	var count int64
	DB := dao.GetMysqlDB()
	DB.Model(&Comment{}).Count(&count)
	return count
}

func CountCommentByPostID(postID uint) int {
	var count int64
	DB := dao.GetMysqlDB()
	DB.Model(&Comment{}).Where("post_id = ?", postID).Count(&count)
	return int(count)
}
