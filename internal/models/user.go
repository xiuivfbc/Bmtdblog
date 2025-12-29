package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"gorm.io/gorm"
)

type User struct {
	ID            uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	Email         string     `gorm:"column:email;type:varchar(255);uniqueIndex;index:idx_email_password_lockstate" json:"email"`
	Telephone     string     `gorm:"column:telephone;type:varchar(20)" json:"telephone"`
	Password      string     `gorm:"column:password;type:varchar(255);index:idx_email_password_lockstate" json:"password"`
	VerifyState   string     `gorm:"column:verify_state;type:varchar(10);default:'0'" json:"verify_state"`
	SecretKey     string     `gorm:"column:secret_key;type:varchar(255)" json:"secret_key"`
	OutTime       time.Time  `gorm:"column:out_time;type:datetime" json:"out_time"`
	GithubLoginId string     `gorm:"column:github_login_id;type:varchar(255);uniqueIndex;default:null" json:"github_login_id"`
	GithubUrl     string     `gorm:"column:github_url;type:varchar(255)" json:"github_url"`
	IsAdmin       bool       `gorm:"column:is_admin;type:bool" json:"is_admin"`
	AvatarUrl     string     `gorm:"column:avatar_url;type:varchar(255)" json:"avatar_url"`
	NickName      string     `gorm:"column:nick_name;type:varchar(50)" json:"nick_name"`
	LockState     bool       `gorm:"column:lock_state;type:bool;index:idx_email_password_lockstate;default:false" json:"lock_state"`
	CreatedAt     *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (user *User) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.Create(user).Error
}

func (user *User) Update() error {
	DB := dao.GetMysqlDB()
	return DB.Model(user).Updates(map[string]any{
		"Email":         user.Email,
		"Telephone":     user.Telephone,
		"Password":      user.Password,
		"VerifyState":   user.VerifyState,
		"SecretKey":     user.SecretKey,
		"OutTime":       user.OutTime,
		"GithubLoginId": user.GithubLoginId,
		"GithubUrl":     user.GithubUrl,
		"IsAdmin":       user.IsAdmin,
		"AvatarUrl":     user.AvatarUrl,
		"NickName":      user.NickName,
		"LockState":     user.LockState,
	}).Error
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	DB := dao.GetMysqlDB()
	err := DB.First(&user, "email = ?", username).Error
	return &user, err
}

func GetUserForLogin(email string) (*User, error) {
	var user User
	DB := dao.GetMysqlDB()
	err := DB.Select("id, email, password, lock_state").
		Where("email = ?", email).
		First(&user).Error
	return &user, err
}

func (user *User) FirstOrCreate() (*User, error) {
	DB := dao.GetMysqlDB()
	err := DB.FirstOrCreate(user, "github_login_id = ?", user.GithubLoginId).Error
	return user, err
}

func IsGithubIdExists(githubId string, id uint) (*User, error) {
	var user User
	DB := dao.GetMysqlDB()
	err := DB.First(&user, "github_login_id = ? and id != ?", githubId, id).Error
	return &user, err
}

func GetUser(id interface{}) (*User, error) {
	var user User
	DB := dao.GetMysqlDB()
	err := DB.First(&user, id).Error
	return &user, err
}

func (user *User) UpdateProfile(avatarUrl, nickName string) error {
	DB := dao.GetMysqlDB()
	return DB.Model(user).Updates(User{AvatarUrl: avatarUrl, NickName: nickName}).Error
}

func (user *User) UpdateEmail(email string) error {
	DB := dao.GetMysqlDB()
	if len(email) > 0 {
		return DB.Model(user).Update("email", email).Error
	} else {
		return DB.Model(user).Update("email", gorm.Expr("NULL")).Error
	}
}

func (user *User) UpdateGithubUserInfo() error {
	var githubLoginId interface{}
	DB := dao.GetMysqlDB()
	if len(user.GithubLoginId) == 0 {
		githubLoginId = gorm.Expr("NULL")
	} else {
		githubLoginId = user.GithubLoginId
	}
	return DB.Model(user).UpdateColumns(map[string]interface{}{
		"github_login_id": githubLoginId,
		"avatar_url":      user.AvatarUrl,
		"github_url":      user.GithubUrl,
	}).Error
}

func (user *User) Lock() error {
	DB := dao.GetMysqlDB()
	return DB.Model(user).UpdateColumns(map[string]interface{}{
		"lock_state": user.LockState,
	}).Error
}

func ListUsers() ([]*User, error) {
	var users []*User
	DB := dao.GetMysqlDB()
	err := DB.Find(&users, "is_admin = ?", false).Error
	return users, err
}
