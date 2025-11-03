package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint       `gorm:"primarykey"`
	CreatedAt     *time.Time `gorm:"autoCreateTime"`
	UpdatedAt     *time.Time `gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"index"`
	Email         string     `gorm:"uniqueindex;index:idx_email_password_lockstate"`
	Telephone     string
	Password      string `gorm:"index:idx_email_password_lockstate"`
	VerifyState   string `gorm:"default:'0'"`
	SecretKey     string
	OutTime       time.Time
	GithubLoginId string `gorm:"uniqueIndex;default:null"`
	GithubUrl     string
	IsAdmin       bool
	AvatarUrl     string
	NickName      string
	LockState     bool `gorm:"index:idx_email_password_lockstate;default:false"`
}

func (user *User) Insert() error {
	return DB.Create(user).Error
}

func (user *User) Update() error {
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
	err := DB.First(&user, "email = ?", username).Error
	return &user, err
}

func GetUserForLogin(email string) (*User, error) {
	var user User
	err := DB.Select("id, email, password, lock_state").
		Where("email = ?", email).
		First(&user).Error
	return &user, err
}

func (user *User) FirstOrCreate() (*User, error) {
	err := DB.FirstOrCreate(user, "github_login_id = ?", user.GithubLoginId).Error
	return user, err
}

func IsGithubIdExists(githubId string, id uint) (*User, error) {
	var user User
	err := DB.First(&user, "github_login_id = ? and id != ?", githubId, id).Error
	return &user, err
}

func GetUser(id interface{}) (*User, error) {
	var user User
	err := DB.First(&user, id).Error
	return &user, err
}

func (user *User) UpdateProfile(avatarUrl, nickName string) error {
	return DB.Model(user).Updates(User{AvatarUrl: avatarUrl, NickName: nickName}).Error
}

func (user *User) UpdateEmail(email string) error {
	if len(email) > 0 {
		return DB.Model(user).Update("email", email).Error
	} else {
		return DB.Model(user).Update("email", gorm.Expr("NULL")).Error
	}
}

func (user *User) UpdateGithubUserInfo() error {
	var githubLoginId interface{}
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
	return DB.Model(user).UpdateColumns(map[string]interface{}{
		"lock_state": user.LockState,
	}).Error
}

func ListUsers() ([]*User, error) {
	var users []*User
	err := DB.Find(&users, "is_admin = ?", false).Error
	return users, err
}
