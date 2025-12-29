package link

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 删除友情链接
// @Description 删除指定ID的友情链接
// @Tags 友情链接管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param id path uint true "友情链接ID"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/link/{id}/delete [post]
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
