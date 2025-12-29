package comment

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 标记评论为已读
// @Description 将指定ID的评论标记为已读状态
// @Tags 评论管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "评论ID"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/comment/{id} [post]
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
