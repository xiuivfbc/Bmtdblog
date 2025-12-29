package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 更新个人资料
// @Description 更新用户个人资料信息
// @Tags 用户管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param avatarUrl formData string false "头像URL"
// @Param nickName formData string false "昵称"
// @Success 200 {object} map[string]interface{} "{"succeed":true,"user":{"avatarUrl":"头像URL","nickName":"昵称"}}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/user/profile [post]
func ProfileUpdate(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	avatarUrl := c.PostForm("avatarUrl")
	nickName := c.PostForm("nickName")
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	log.Debug("ProfileUpdate", zap.String("avatarUrl", avatarUrl), zap.String("nickName", nickName), zap.Uint("userId", user.ID))
	err = user.UpdateProfile(avatarUrl, nickName)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["user"] = models.User{AvatarUrl: avatarUrl, NickName: nickName}
}
