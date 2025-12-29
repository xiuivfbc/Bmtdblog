package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 绑定邮箱
// @Description 绑定用户邮箱
// @Tags 用户管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param email formData string true "邮箱地址"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/user/bindemail [post]
func BindEmail(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	email := c.PostForm("email")
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	log.Debug("BindEmail", zap.String("email", email), zap.Uint("userId", user.ID))
	if len(user.Email) > 0 {
		res["message"] = "email have bound"
		return
	}
	_, err = models.GetUserByUsername(email)
	if err == nil {
		res["message"] = "email have be registered"
		return
	}
	err = user.UpdateEmail(email)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
