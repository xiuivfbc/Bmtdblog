package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Subscriber struct {
	ID             uint       `gorm:"column:id;type:uint;primary_key;AUTO_INCREMENT" json:"id"`
	Email          string     `gorm:"column:email;type:varchar(255);uniqueIndex" json:"email"`
	VerifyState    bool       `gorm:"column:verify_state;type:bool;default:false" json:"verify_state"`
	SubscribeState bool       `gorm:"column:subscribe_state;type:bool;default:true" json:"subscribe_state"`
	OutTime        time.Time  `gorm:"column:out_time;type:datetime;default:null" json:"out_time"`
	SecretKey      string     `gorm:"column:secret_key;type:varchar(255)" json:"secret_key"`
	Signature      string     `gorm:"column:signature;type:varchar(255)" json:"signature"`
	CreatedAt      *time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (s *Subscriber) Insert() error {
	DB := dao.GetMysqlDB()
	return DB.FirstOrCreate(s, "email = ?", s.Email).Error
}

func (s *Subscriber) Update() error {
	DB := dao.GetMysqlDB()
	return DB.Model(s).UpdateColumns(map[string]interface{}{
		"verify_state":    s.VerifyState,
		"subscribe_state": s.SubscribeState,
		"out_time":        s.OutTime,
		"signature":       s.Signature,
		"secret_key":      s.SecretKey,
	}).Error
}

func ListSubscriber(valid bool) ([]*Subscriber, error) {
	var subscribers []*Subscriber
	DB := dao.GetMysqlDB()
	db := DB.Model(&Subscriber{})
	if valid {
		db.Where("verify_state = ? and subscribe_state = ?", true, true)
	}
	err := db.Find(&subscribers).Error
	return subscribers, err
}

func CountSubscriber() (int64, error) {
	var count int64
	DB := dao.GetMysqlDB()
	err := DB.Model(&Subscriber{}).Where("verify_state = ? and subscribe_state = ?", true, true).Count(&count).Error
	return count, err
}

func GetSubscriberByEmail(mail string) (*Subscriber, error) {
	var subscriber Subscriber
	DB := dao.GetMysqlDB()
	err := DB.First(&subscriber, "email = ?", mail).Error
	return &subscriber, err
}

func GetSubscriberBySignature(key string) (*Subscriber, error) {
	var subscriber Subscriber
	DB := dao.GetMysqlDB()
	err := DB.First(&subscriber, "signature = ?", key).Error
	return &subscriber, err
}

func GetSubscriberById(id uint) (*Subscriber, error) {
	var subscriber Subscriber
	DB := dao.GetMysqlDB()
	err := DB.First(&subscriber, id).Error
	return &subscriber, err
}
