package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 锁定/解锁用户
// @Description 切换指定ID用户的锁定状态（锁定/解锁）
// @Tags 用户管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "用户ID"
// @Success 200 {object} map[string]interface{} "{\"succeed\":true}"
// @Failure 200 {object} map[string]interface{} "{\"succeed\":false,\"message\":\"错误信息\"}"
// @Router /admin/user/{id}/lock [post]
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
