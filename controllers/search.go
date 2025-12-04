package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
)

// SearchGet 搜索页面
func SearchGet(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	tags := c.QueryArray("tags")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}

	sortBy := c.DefaultQuery("sort", "relevance")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	pageSize := 10

	// 构建搜索请求
	req := &models.SearchRequest{
		Query:    keyword,
		Tags:     tags,
		Page:     page,
		Size:     pageSize,
		SortBy:   sortBy,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	// 执行搜索
	fmt.Printf("开始搜索: keyword=%s, page=%d, sort=%s\n", keyword, page, sortBy)
	system.LogInfo(c, "开始搜索", "keyword", keyword, "page", page, "sort", sortBy)
	searchResp, err := models.SearchPosts(req)
	if err != nil {
		system.LogError(c, "搜索失败", "error", err, "keyword", keyword)
		c.HTML(http.StatusOK, "search/results.html", gin.H{
			"keyword": keyword,
			"error":   "搜索服务暂时不可用，请稍后重试",
			"user":    c.MustGet(ContextUserKey),
			"cfg":     system.GetConfiguration(),
		})
		return
	}

	// 记录搜索日志（用于分析热门搜索词）
	go recordSearchLog(keyword, int(searchResp.Total))

	system.LogInfo(c, "搜索完成", "keyword", keyword, "results", len(searchResp.Posts), "total", searchResp.Total)

	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusOK, "search/results.html", gin.H{
		"keyword":         keyword,
		"posts":           searchResp.Posts,
		"total":           searchResp.Total,
		"page":            page,
		"totalPages":      int(math.Ceil(float64(searchResp.Total) / float64(pageSize))),
		"took":            searchResp.Took,
		"sortBy":          sortBy,
		"allTags":         models.MustListTag(),
		"selectedTags":    tags,
		"archives":        models.MustListPostArchives(),
		"links":           models.MustListLinks(),
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
		"dateFrom":        dateFrom,
		"dateTo":          dateTo,
		"suggestions":     searchResp.Suggestions,
		"user":            user,
		"cfg":             system.GetConfiguration(),
	})
}

// SearchIndexGet 搜索首页
func SearchIndexGet(c *gin.Context) {
	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusOK, "search/index.html", gin.H{
		"user":    user,
		"allTags": models.MustListTag(),
		"cfg":     system.GetConfiguration(),
	})
}

// SearchSuggestionsAPI 搜索建议API
func SearchSuggestionsAPI(c *gin.Context) {
	prefix := strings.TrimSpace(c.Query("q"))
	if len(prefix) < 2 {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	suggestions, err := models.GetSearchSuggestions(prefix, 10)
	if err != nil {
		system.LogError(c, "获取搜索建议失败", "error", err, "prefix", prefix)
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// recordSearchLog 记录搜索日志
func recordSearchLog(keyword string, resultCount int) {
	if keyword == "" {
		return
	}

	// 这里可以实现搜索日志记录
	// 比如记录到数据库或日志文件，用于分析热门搜索词
	system.Logger.Info("搜索记录",
		"keyword", keyword,
		"result_count", resultCount,
		"timestamp", time.Now())
}
