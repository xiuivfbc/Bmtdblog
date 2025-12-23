package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

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
