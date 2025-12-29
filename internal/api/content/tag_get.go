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
)

// @Summary 标签页内容
// @Description 获取指定标签下的文章列表，支持分页
// @Tags 内容浏览
// @Accept html
// @Produce html
// @Param tag path string true "标签名称"
// @Param page query int false "页码，默认1"
// @Success 200 {html} string "标签文章列表页面"
// @Failure 500 {string} string "服务器内部错误"
// @Router /tag/{tag} [get]
func TagGet(c *gin.Context) {
	var (
		tagName   string
		page      string
		pageIndex int
		pageSize  = config.GetConfiguration().PageSize
		total     int
		err       error
		policy    *bluemonday.Policy
		posts     []*models.Post
	)
	tagName = c.Param("tag")
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListPublishedPost(tagName, pageIndex, pageSize)
	if err != nil {
		log.Error("models.ListPublishedPost error", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByTag(tagName)
	if err != nil {
		log.Error("models.CountPostByTag error", "err", err)
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
