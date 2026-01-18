package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/internal/api/dao"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// @Summary 本地备份
// @Description 备份数据库到本地文件系统
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "{"succeed":true,"message":"备份成功"}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/backup/local [post]
// 保持LocalBackup函数不变，让它调用新的Backup函数
func LocalBackup(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)

	log.Debug("LocalBackup")
	err = Backup()
	if err != nil {
		res["message"] = err.Error()
		res["succeed"] = false
		return
	}

	res["succeed"] = true
	res["message"] = "本地备份成功"
}

// localBackup 执行本地备份操作
// 将localBackup改为Backup（大写开头，公开函数）
func Backup() (err error) {
	conf := config.GetConfiguration()

	// 检查备份功能是否启用
	if !conf.Backup.Enabled {
		err = errors.New("备份功能未启用")
		return
	}

	log.Debug("开始本地备份...")

	// 使用mysqldump生成备份内容
	bodyBytes, err := dao.Mysqldump(conf.Mysql.Host, conf.Mysql.Port, conf.Mysql.User, conf.Mysql.Password, conf.Mysql.DbName)
	if err != nil {
		log.Error("mysqldump错误", "err", err)
		return
	}

	// 如果配置了备份密钥，则加密备份内容
	if len(conf.Backup.BackupKey) > 0 {
		bodyBytes, err = common.Encrypt(bodyBytes, []byte(conf.Backup.BackupKey))
		if err != nil {
			log.Error("加密备份文件错误", "err", err)
			return
		}
	}

	// 确保备份目录存在
	backupDir := conf.Backup.LocalPath
	if err = os.MkdirAll(backupDir, 0755); err != nil {
		log.Error("创建备份目录失败", "err", err)
		err = errors.Wrap(err, "创建备份目录失败")
		return
	}

	// 生成备份文件名
	fileExt := ".sql"
	if len(conf.Backup.BackupKey) > 0 {
		fileExt = ".sql.encrypted"
	}
	fileName := fmt.Sprintf("Bmtdblog_%s%s", common.GetCurrentTime().Format("20060102150405"), fileExt)
	filePath := filepath.Join(backupDir, fileName)

	// 写入备份文件
	if err = os.WriteFile(filePath, bodyBytes, 0644); err != nil {
		log.Error("写入备份文件失败", "err", err)
		err = errors.Wrap(err, "写入备份文件失败")
		return
	}

	log.Debug("本地备份成功", "file", filePath)
	return nil
}
