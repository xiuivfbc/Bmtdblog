package content

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 搜索文章
// @Description 根据关键词搜索文章，支持标签过滤、时间范围过滤和排序
// @Tags 搜索功能
// @Accept json
// @Produce html
// @Param q query string true "搜索关键词"
// @Param tags query []string false "标签过滤"
// @Param page query int false "页码，默认1"
// @Param sort query string false "排序方式：relevance(相关性)、date(日期)、read(阅读量)"
// @Param date_from query string false "开始日期(YYYY-MM-DD)"
// @Param date_to query string false "结束日期(YYYY-MM-DD)"
// @Success 200 {string} html "搜索结果页面"
// @Failure 200 {string} html "搜索失败页面"
// @Router /search [get]
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
	log.Debug("开始搜索", "keyword", keyword, "page", page, "sort", sortBy)
	searchResp, err := models.SearchPosts(req)
	if err != nil {
		log.Error("搜索失败", "error", err, "keyword", keyword)
		c.HTML(http.StatusOK, "search/results.html", gin.H{
			"keyword": keyword,
			"error":   "搜索服务暂时不可用，请稍后重试",
			"user":    c.MustGet(common.ContextUserKey),
			"cfg":     config.GetConfiguration(),
		})
		return
	}

	// 记录搜索日志（用于分析热门搜索词）
	go recordSearchLog(keyword, int(searchResp.Total))

	log.Debug("搜索完成", "keyword", keyword, "results", len(searchResp.Posts), "total", searchResp.Total)

	user, _ := c.Get(common.ContextUserKey)
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
		"cfg":             config.GetConfiguration(),
	})
}
