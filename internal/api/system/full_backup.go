package system

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/xiuivfbc/bmtdblog/internal/api/upload"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func Backup(c ...*gin.Context) (err error) {
	var (
		ret       upload.PutRet
		bodyBytes []byte
		cfg       = system.GetConfiguration()
	)
	if !cfg.Backup.Enabled || !cfg.Qiniu.Enabled {
		err = errors.New("backup or qiniu not enabled")
		return
	}

	// 日志输出（支持有无上下文）
	if len(c) > 0 && c[0] != nil {
		system.LogDebug(c[0], "start backup...")
	} else {
		system.Logger.Debug("start backup...")
	}

	dsn := cfg.Database.DSN
	host, port, username, password, database, err := parseMySQLDSN(dsn)
	if err != nil {
		system.Logger.Error("parse mysql dsn error", "err", err)
		return
	}
	bodyBytes, err = mysqldump(host, port, username, password, database)
	if err != nil {
		system.Logger.Error("mysqldump error", "err", err)
		return
	}
	if len(cfg.Backup.BackupKey) > 0 {
		bodyBytes, err = common.Encrypt(bodyBytes, []byte(cfg.Backup.BackupKey))
		if err != nil {
			system.Logger.Error("encrypt backup file error", "err", err)
			return
		}
	}
	// upload to qiniu
	putPolicy := storage.PutPolicy{
		Scope: cfg.Qiniu.Bucket,
	}
	mac := qbox.NewMac(cfg.Qiniu.AccessKey, cfg.Qiniu.SecretKey)
	token := putPolicy.UploadToken(mac)
	uploader := storage.NewFormUploader(&storage.Config{})
	putExtra := storage.PutExtra{}
	fileName := fmt.Sprintf("Bmtdblog_%s.sql", common.GetCurrentTime().Format("20060102150405"))
	err = uploader.Put(context.Background(), &ret, token, fileName, bytes.NewReader(bodyBytes), int64(len(bodyBytes)), &putExtra)
	if err != nil {
		system.Logger.Debug("backup error", "err", err)
		return
	}
	system.Logger.Debug("backup successfully.")
	return err
}
