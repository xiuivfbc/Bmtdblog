package comment

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 标记所有评论为已读
// @Description 将所有未读评论标记为已读状态
// @Tags 评论管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/read_all [post]
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
