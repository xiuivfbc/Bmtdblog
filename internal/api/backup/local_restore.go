package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 本地恢复
// @Description 从本地备份文件恢复数据库
// @Tags 备份管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "备份文件"
// @Param encrypted formData bool false "是否加密"
// @Success 200 {object} map[string]interface{} "{"succeed":true,"message":"恢复成功"}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/restore [post]
func LocalRestore(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)

	log.Debug("LocalRestore")

	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		res["message"] = "获取上传文件失败: " + err.Error()
		res["succeed"] = false
		return
	}
	defer file.Close()

	// 检查备份功能是否启用
	conf := config.GetConfiguration()
	if !conf.Backup.Enabled {
		res["message"] = "备份功能未启用"
		res["succeed"] = false
		return
	}

	// 检查是否需要解密
	encrypted := c.PostForm("encrypted") == "true"
	if encrypted && len(conf.Backup.BackupKey) == 0 {
		res["message"] = "备份文件已加密，但未配置备份密钥"
		res["succeed"] = false
		return
	}

	// 读取文件内容
	bodyBytes, err := io.ReadAll(file)
	if err != nil {
		res["message"] = "读取备份文件失败: " + err.Error()
		res["succeed"] = false
		return
	}

	// 如果文件是加密的，进行解密
	if encrypted {
		bodyBytes, err = common.Decrypt(bodyBytes, []byte(conf.Backup.BackupKey))
		if err != nil {
			res["message"] = "解密备份文件失败: " + err.Error()
			res["succeed"] = false
			return
		}
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Mysql.User,
		conf.Mysql.Password,
		conf.Mysql.Host,
		conf.Mysql.Port,
		conf.Mysql.DbName,
	)

	// 执行恢复操作
	err = mysqlRestore(string(bodyBytes), dsn)
	if err != nil {
		res["message"] = "恢复数据库失败: " + err.Error()
		res["succeed"] = false
		return
	}

	res["succeed"] = true
	res["message"] = "数据库恢复成功"
}

// @Summary 获取本地备份文件列表
// @Description 获取本地备份目录中的所有备份文件
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true,"files":["file1.sql","file2.sql"]}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/local/files [get]
func GetLocalBackupFiles(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)

	conf := config.GetConfiguration()
	backupDir := conf.Backup.LocalPath

	// 检查备份目录是否存在
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		res["succeed"] = true
		res["files"] = []string{}
		return
	}

	// 读取目录中的文件
	files, err := os.ReadDir(backupDir)
	if err != nil {
		res["message"] = "读取备份目录失败: " + err.Error()
		res["succeed"] = false
		return
	}

	// 过滤出.sql和.sql.encrypted文件
	var backupFiles []string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			if filepath.Ext(name) == ".sql" || filepath.Ext(name) == ".encrypted" {
				backupFiles = append(backupFiles, name)
			}
		}
	}

	res["succeed"] = true
	res["files"] = backupFiles
}

// @Summary 从指定文件名恢复
// @Description 根据文件名从本地备份目录恢复数据库
// @Tags 备份管理
// @Accept json
// @Produce json
// @Param filename query string true "备份文件名"
// @Success 200 {object} map[string]interface{} "{"succeed":true,"message":"恢复成功"}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/restore/file [post]
func RestoreFromLocalFile(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)

	log.Debug("RestoreFromLocalFile")

	// 获取文件名参数
	filename := c.Query("filename")
	if filename == "" {
		res["message"] = "文件名不能为空"
		res["succeed"] = false
		return
	}

	conf := config.GetConfiguration()

	// 检查备份功能是否启用
	if !conf.Backup.Enabled {
		res["message"] = "备份功能未启用"
		res["succeed"] = false
		return
	}

	// 构建文件路径
	backupDir := conf.Backup.LocalPath
	filePath := filepath.Join(backupDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		res["message"] = "备份文件不存在"
		res["succeed"] = false
		return
	}

	// 读取文件内容
	bodyBytes, err := os.ReadFile(filePath)
	if err != nil {
		res["message"] = "读取备份文件失败: " + err.Error()
		res["succeed"] = false
		return
	}

	// 检查是否需要解密
	encrypted := filepath.Ext(filename) == ".encrypted"
	if encrypted {
		if len(conf.Backup.BackupKey) == 0 {
			res["message"] = "备份文件已加密，但未配置备份密钥"
			res["succeed"] = false
			return
		}

		// 解密文件内容
		bodyBytes, err = common.Decrypt(bodyBytes, []byte(conf.Backup.BackupKey))
		if err != nil {
			res["message"] = "解密备份文件失败: " + err.Error()
			res["succeed"] = false
			return
		}
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Mysql.User,
		conf.Mysql.Password,
		conf.Mysql.Host,
		conf.Mysql.Port,
		conf.Mysql.DbName,
	)

	// 执行恢复操作
	err = mysqlRestore(string(bodyBytes), dsn)
	if err != nil {
		res["message"] = "恢复数据库失败: " + err.Error()
		res["succeed"] = false
		return
	}

	res["succeed"] = true
	res["message"] = "数据库恢复成功"
}
