package backup

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

func RestorePost(c *gin.Context) {
	var (
		fileName  string
		fileUrl   string
		err       error
		res       = gin.H{}
		resp      *http.Response
		bodyBytes []byte
		cfg       = config.GetConfiguration()
	)
	defer common.WriteJSON(c, res)
	fileName = c.PostForm("fileName")
	if fileName == "" {
		res["message"] = "fileName cannot be empty."
		return
	}
	if !cfg.Backup.Enabled || !cfg.Qiniu.Enabled {
		res["message"] = "backup or qiniu not enabled"
		return
	}

	fileUrl = cfg.Qiniu.FileServer + fileName
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
	if len(cfg.Backup.BackupKey) > 0 {
		bodyBytes, err = common.Decrypt(bodyBytes, []byte(cfg.Backup.BackupKey))
		if err != nil {
			res["message"] = err.Error()
			return
		}
	}

	err = mysqlRestore(string(bodyBytes), cfg.Database.Dsn)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
