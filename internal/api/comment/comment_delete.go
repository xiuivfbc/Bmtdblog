package comment

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

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
