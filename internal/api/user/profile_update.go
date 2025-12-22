package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

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
