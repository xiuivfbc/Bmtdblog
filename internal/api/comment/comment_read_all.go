package comment

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func CommentReadAll(c *gin.Context) {
	log.Debug("CommentReadAll")
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	err = models.SetAllCommentRead()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
