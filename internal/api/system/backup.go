package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/xiuivfbc/bmtdblog/internal/api/upload"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func BackupPost(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	err = Backup(c)
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

	err = mysqlRestore(string(bodyBytes), cfg.Database.DSN)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

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

func parseMySQLDSN(dsn string) (host, port, username, password, database string, err error) {
	re := regexp.MustCompile(`^([^:]+):([^@]*)@tcp\(([^:]+):(\d+)\)/([^?]+)`)
	matches := re.FindStringSubmatch(dsn)

	if len(matches) != 6 {
		err = errors.New("invalid mysql dsn format")
		return
	}

	username = matches[1]
	password = matches[2]
	host = matches[3]
	port = matches[4]
	database = matches[5]
	return
}

func mysqldump(host, port, username, password, database string) ([]byte, error) {
	// 构造 mysqldump 命令
	cmd := exec.Command("mysqldump",
		"-h", host,
		"-P", port,
		"-u", username,
		fmt.Sprintf("-p%s", password),
		"--single-transaction",
		"--routines",
		"--triggers",
		database,
	)

	// 执行命令并获取输出
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "mysqldump command failed")
	}

	return output, nil
}

func mysqlRestore(sqlContent string, dsn string) error {
	host, port, username, password, database, err := parseMySQLDSN(dsn)
	if err != nil {
		return err
	}
	cmd := exec.Command("mysql",
		"-h", host,
		"-P", port,
		"-u", username,
		fmt.Sprintf("-p%s", password),
		database,
	)
	cmd.Stdin = strings.NewReader(sqlContent)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "mysql restore failed: %s", string(output))
	}
	return nil
}
