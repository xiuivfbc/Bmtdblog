package content

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func PostDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err := common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	log.Debug("PostDelete", zap.Uint("id", id))
	post := &models.Post{}
	post.ID = id
	err = post.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	models.DeletePostTagByPostId(id)
	res["succeed"] = true
}
