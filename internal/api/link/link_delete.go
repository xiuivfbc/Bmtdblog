package link

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func LinkDelete(c *gin.Context) {
	var (
		err error
		id  uint
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	log.Debug("LinkDelete", zap.Uint("id", id))

	link := new(models.Link)
	link.ID = id
	err = link.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
