package upload

import (
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

type Uploader interface {
	upload(file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}

func Upload(c *gin.Context) {
	var (
		err      error
		res      = gin.H{}
		url      string
		uploader Uploader
		file     multipart.File
		fh       *multipart.FileHeader
		cfg      = config.GetConfiguration()
	)
	defer common.WriteJSON(c, res)
	file, fh, err = c.Request.FormFile("file")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	if cfg.FileServer == "smms" && cfg.Smms.Enabled {
		uploader = SmmsUploader{}
	}
	if cfg.FileServer == "qiniu" && cfg.Qiniu.Enabled {
		uploader = QiniuUploader{}
	}
	if uploader == nil {
		res["message"] = "no available fileserver"
		return
	}
	url, err = uploader.upload(file, fh)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
	res["url"] = url
}
