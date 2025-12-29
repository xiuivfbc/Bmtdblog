package content

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 归档页内容
// @Description 获取指定年份和月份的文章列表，支持分页
// @Tags 内容浏览
// @Accept html
// @Produce html
// @Param year path string true "年份"
// @Param month path string true "月份"
// @Param page query int false "页码，默认1"
// @Success 200 {html} string "归档文章列表页面"
// @Failure 500 {string} string "服务器内部错误"
// @Router /archives/{year}/{month} [get]
func ArchiveGet(c *gin.Context) {
	var (
		year      string
		month     string
		page      string
		pageIndex int
		pageSize  = config.GetConfiguration().PageSize
		total     int
		err       error
		posts     []*models.Post
		policy    *bluemonday.Policy
	)
	year = c.Param("year")
	month = c.Param("month")
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	log.Debug("ArchiveGet", zap.String("year", year), zap.String("month", month), zap.Int("pageIndex", pageIndex))
	posts, err = models.ListPostByArchive(year, month, pageIndex, pageSize)
	if err != nil {
		log.Error("models.ListPostByArchive err", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByArchive(year, month)
	if err != nil {
		log.Error("models.CountPostByArchive err", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	policy = bluemonday.StrictPolicy()
	for _, post := range posts {
		post.Tags, _ = models.ListTagByPostId(post.ID)
		post.Body = policy.Sanitize(string(blackfriday.MarkdownCommon([]byte(post.Body))))
		post.CommentTotal = models.CountCommentByPostID(post.ID)
	}
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "index/index.html", gin.H{
		"posts":           posts,
		"tags":            models.MustListTag(),
		"archives":        models.MustListPostArchives(),
		"links":           models.MustListLinks(),
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
		"user":            user,
		"cfg":             config.GetConfiguration(),
	})

}
