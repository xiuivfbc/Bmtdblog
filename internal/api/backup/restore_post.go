package backup

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"go.uber.org/zap"
)

// @Summary 恢复文章
// @Description 从七牛云恢复文章数据
// @Tags 备份管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param fileName formData string true "备份文件名"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/restore [post]
func RestorePost(c *gin.Context) {
	var (
		fileName  string
		fileUrl   string
		err       error
		res       = gin.H{}
		resp      *http.Response
		bodyBytes []byte
		conf      = config.GetConfiguration()
	)
	defer common.WriteJSON(c, res)
	fileName = c.PostForm("fileName")
	log.Debug("RestorePost", zap.String("fileName", fileName))
	if fileName == "" {
		res["message"] = "fileName cannot be empty."
		return
	}
	if !conf.Backup.Enabled || !conf.Qiniu.Enabled {
		res["message"] = "backup or qiniu not enabled"
		return
	}

	fileUrl = conf.Qiniu.FileServer + fileName
	resp, err = http.Get(fileUrl)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	defer resp.Body.Close()
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	if len(conf.Backup.BackupKey) > 0 {
		bodyBytes, err = common.Decrypt(bodyBytes, []byte(conf.Backup.BackupKey))
		if err != nil {
			res["message"] = err.Error()
			return
		}
	}
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=UTC",
		conf.Mysql.User, conf.Mysql.Password, conf.Mysql.Host, conf.Mysql.Port, conf.Mysql.DbName,
	)

	err = mysqlRestore(string(bodyBytes), dsn)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
