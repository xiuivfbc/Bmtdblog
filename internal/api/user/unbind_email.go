package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 解绑邮箱
// @Description 解绑用户的邮箱地址
// @Tags 个人资料管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/profile/email/unbind [post]
func UnbindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	if user.Email == "" {
		res["message"] = "email haven't bound"
		return
	}
	err = user.UpdateEmail("")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
