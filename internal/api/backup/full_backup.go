package backup

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"github.com/xiuivfbc/bmtdblog/internal/api/upload"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

func Backup(c ...*gin.Context) (err error) {
	var (
		ret       upload.PutRet
		bodyBytes []byte
		cfg       = config.GetConfiguration()
	)
	if !cfg.Backup.Enabled || !cfg.Qiniu.Enabled {
		err = errors.New("backup or qiniu not enabled")
		return
	}

	// 日志输出（支持有无上下文）
	if len(c) > 0 && c[0] != nil {
		log.Debug("start backup...")
	} else {
		log.Debug("start backup...")
	}

	dsn := cfg.Database.Dsn
	host, port, username, password, database, err := dao.ParseMySQLDSN(dsn)
	if err != nil {
		log.Error("parse mysql dsn error", "err", err)
		return
	}
	bodyBytes, err = dao.Mysqldump(host, port, username, password, database)
	if err != nil {
		log.Error("mysqldump error", "err", err)
		return
	}
	if len(cfg.Backup.BackupKey) > 0 {
		bodyBytes, err = common.Encrypt(bodyBytes, []byte(cfg.Backup.BackupKey))
		if err != nil {
			log.Error("encrypt backup file error", "err", err)
			return
		}
	}
	// upload to qiniu
	putPolicy := storage.PutPolicy{
		Scope: cfg.Qiniu.Bucket,
	}
	mac := qbox.NewMac(cfg.Qiniu.Accesskey, cfg.Qiniu.Secretkey)
	token := putPolicy.UploadToken(mac)
	uploader := storage.NewFormUploader(&storage.Config{})
	putExtra := storage.PutExtra{}
	fileName := fmt.Sprintf("Bmtdblog_%s.sql", common.GetCurrentTime().Format("20060102150405"))
	err = uploader.Put(context.Background(), &ret, token, fileName, bytes.NewReader(bodyBytes), int64(len(bodyBytes)), &putExtra)
	if err != nil {
		log.Debug("backup error", "err", err)
		return
	}
	log.Debug("backup successfully.")
	return err
}
