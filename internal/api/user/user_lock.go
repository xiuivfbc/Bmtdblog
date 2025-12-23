package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func UserLock(c *gin.Context) {
	var (
		err  error
		id   uint
		res  = gin.H{}
		user *models.User
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	log.Debug("UserLock", zap.Uint("id", id))
	user, err = models.GetUser(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	user.LockState = !user.LockState
	err = user.Lock()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
