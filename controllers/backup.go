package controllers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/wangsongyan/wblog/helpers"
	"github.com/wangsongyan/wblog/system"
)

func BackupPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer writeJSON(c, res)
	err = Backup()
	if err != nil {
		res["message"] = err.Error()
	} else {
		res["succeed"] = true
	}
}

func RestorePost(c *gin.Context) {
	var (
		fileName  string
		fileUrl   string
		err       error
		res       = gin.H{}
		resp      *http.Response
		bodyBytes []byte
		cfg       = system.GetConfiguration()
	)
	defer writeJSON(c, res)
	fileName = c.PostForm("fileName")
	if fileName == "" {
		res["message"] = "fileName cannot be empty."
		return
	}

	if !cfg.Backup.Enabled || !cfg.Qiniu.Enabled {
		res["message"] = "backup or quniu not enabled"
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
		bodyBytes, err = helpers.Decrypt(bodyBytes, []byte(cfg.Backup.BackupKey))
		if err != nil {
			res["message"] = err.Error()
			return
		}
	}
	err = os.WriteFile(fileName, bodyBytes, os.ModePerm)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func Backup() (err error) {
	var (
		u         *url.URL
		exist     bool
		ret       PutRet
		bodyBytes []byte
		cfg       = system.GetConfiguration()
	)

	if !cfg.Backup.Enabled || !cfg.Qiniu.Enabled {
		err = errors.New("backup or qiniu not enabled")
		return
	}

	u, err = url.Parse(cfg.Database.DSN)
	if err != nil {
		system.Logger.Debug("parse dsn error", "err", err)
		return
	}
	exist, _ = helpers.PathExists(u.Path)
	if !exist {
		err = errors.New("database file doesn't exists.")
		system.Logger.Debug("database file doesn't exists", "err", err)
		return
	}
	system.Logger.Debug("start backup...")
	bodyBytes, err = os.ReadFile(u.Path)
	if err != nil {
		system.Logger.Error("read database file error", "err", err)
		return
	}
	if len(cfg.Backup.BackupKey) > 0 {
		bodyBytes, err = helpers.Encrypt(bodyBytes, []byte(cfg.Backup.BackupKey))
		if err != nil {
			system.Logger.Error("encrypt backup file error", "err", err)
			return
		}
	}

	putPolicy := storage.PutPolicy{
		Scope: cfg.Qiniu.Bucket,
	}
	mac := qbox.NewMac(cfg.Qiniu.AccessKey, cfg.Qiniu.SecretKey)
	token := putPolicy.UploadToken(mac)
	uploader := storage.NewFormUploader(&storage.Config{})
	putExtra := storage.PutExtra{}

	fileName := fmt.Sprintf("wblog_%s.db", helpers.GetCurrentTime().Format("20060102150405"))
	err = uploader.Put(context.Background(), &ret, token, fileName, bytes.NewReader(bodyBytes), int64(len(bodyBytes)), &putExtra)
	if err != nil {
		system.Logger.Debug("backup error", "err", err)
		return
	}
	system.Logger.Debug("backup successfully.")
	return err
}
