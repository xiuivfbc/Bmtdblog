package interaction

import (
	"fmt"
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func sendActiveEmail(subscriber *models.Subscriber) (err error) {
	uuid := common.UUID()
	duration, _ := time.ParseDuration("30m")
	subscriber.OutTime = common.GetCurrentTime().Add(duration)
	subscriber.SecretKey = uuid
	signature := common.Md5(subscriber.Email + uuid + subscriber.OutTime.Format("20060102150405"))
	subscriber.Signature = signature
	err = common.SendMail(subscriber.Email, fmt.Sprintf("[%s]邮箱验证", system.GetConfiguration().Title), fmt.Sprintf("%s/active?sid=%s", system.GetConfiguration().Domain, signature))
	if err != nil {
		return
	}
	err = subscriber.Update()
	return
}
