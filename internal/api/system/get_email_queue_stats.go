package system

import "github.com/xiuivfbc/bmtdblog/internal/system"

func getEmailQueueStats() (map[string]interface{}, error) {
	if system.EmailQueueInstance == nil {
		return map[string]interface{}{
			"status":      "disabled",
			"workers":     0,
			"queue_size":  0,
			"failed_size": 0,
		}, nil
	}

	return system.EmailQueueInstance.GetQueueStats()
}
