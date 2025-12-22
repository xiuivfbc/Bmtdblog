package content

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"go.uber.org/zap"
)

// recordSearchLog 记录搜索日志
func recordSearchLog(keyword string, resultCount int) {
	if keyword == "" {
		return
	}

	// 这里可以实现搜索日志记录
	// 比如记录到数据库或日志文件，用于分析热门搜索词
	log.Info("搜索记录",
		zap.String("keyword", keyword),
		zap.Int("result_count", resultCount),
		zap.Time("timestamp", time.Now()))
}
