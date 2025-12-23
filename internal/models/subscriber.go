package models

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

type Subscriber struct {
	ID             uint       `gorm:"primarykey"`
	CreatedAt      *time.Time `gorm:"autoCreateTime"`
	UpdatedAt      *time.Time `gorm:"autoUpdateTime"`
	DeletedAt      *time.Time `gorm:"index"`
	Email          string     `gorm:"type:varchar(255);uniqueIndex"`
	VerifyState    bool       `gorm:"default:false"`
	SubscribeState bool       `gorm:"default:true"`
	OutTime        time.Time  `gorm:"default:null"`
	SecretKey      string
	Signature      string
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
