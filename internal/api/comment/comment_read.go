package comment

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func CommentRead(c *gin.Context) {
	var (
		id  uint
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	log.Debug("CommentRead", zap.Uint("id", id))
	comment := new(models.Comment)
	comment.ID = id
	err = comment.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
