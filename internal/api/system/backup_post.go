package system

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
)

func BackupPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	err = Backup(c)
	if err != nil {
		res["message"] = err.Error()
	} else {
		res["succeed"] = true
	}
}
