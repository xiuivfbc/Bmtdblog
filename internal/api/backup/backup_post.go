package backup

import (
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
)

// @Summary 备份文章
// @Description 备份文章数据到七牛云
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/post [post]
func BackupPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	log.Debug("BackupPost")
	err = Backup(c)
	if err != nil {
		res["message"] = err.Error()
	} else {
		res["succeed"] = true
	}
}
