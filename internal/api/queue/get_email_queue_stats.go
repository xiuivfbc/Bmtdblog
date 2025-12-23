package queue

import (
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
)

func getEmailQueueStats() (map[string]interface{}, error) {
	if dao.EmailQueueInstance == nil {
		return map[string]interface{}{
			"status":      "disabled",
			"workers":     0,
			"queue_size":  0,
			"failed_size": 0,
		}, nil
	}

	return dao.EmailQueueInstance.GetQueueStats()
}
