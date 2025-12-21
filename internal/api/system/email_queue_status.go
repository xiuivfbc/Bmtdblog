package system

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

func EmailQueueStatus(c *gin.Context) {
	var (
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	stats, err := getEmailQueueStats()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["stats"] = stats
}
