package content

import (
	"time"

	"github.com/xiuivfbc/bmtdblog/internal/config"
)

// recordSearchLog 记录搜索日志
func recordSearchLog(keyword string, resultCount int) {
	if keyword == "" {
		return
	}

	// 这里可以实现搜索日志记录
	// 比如记录到数据库或日志文件，用于分析热门搜索词
	config.Logger.Info("搜索记录",
		"keyword", keyword,
		"result_count", resultCount,
		"timestamp", time.Now())
}
