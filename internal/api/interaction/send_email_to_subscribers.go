package interaction

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func sendEmailToSubscribers(subject, body string) (err error) {
	var (
		subscribers []*models.Subscriber
		emails      = make([]string, 0)
	)
	subscribers, err = models.ListSubscriber(true)
	if err != nil {
		return
	}
	for _, subscriber := range subscribers {
		emails = append(emails, subscriber.Email)
	}
	if len(emails) == 0 {
		err = errors.New("no subscribers!")
		return
	}
	err = common.SendMail(strings.Join(emails, ";"), subject, body)
	return
}
