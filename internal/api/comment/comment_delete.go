package comment

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 删除评论
// @Description 删除自己发布的评论
// @Tags 评论管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "评论ID"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /visitor/comment/{id}/delete [post]
func CommentDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		cid uint
	)
	defer common.WriteJSON(c, res)

	s := sessions.Default(c)
	userId := s.Get(common.SessionKey).(uint)

	cid, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	log.Debug("CommentDelete", zap.Uint("cid", cid), zap.Uint("userId", userId))
	comment := &models.Comment{
		UserID: userId,
	}
	comment.ID = cid
	err = comment.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
