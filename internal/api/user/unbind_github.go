package user

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 解绑Github账号
// @Description 解绑用户的Github账号
// @Tags 个人资料管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/profile/github/unbind [post]
func UnbindGithub(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	sessionUser, _ := c.Get(common.ContextUserKey)
	user := sessionUser.(*models.User)
	if user.GithubLoginId == "" {
		res["message"] = "github haven't bound"
		return
	}
	user.GithubLoginId = ""
	err = user.UpdateGithubUserInfo()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
