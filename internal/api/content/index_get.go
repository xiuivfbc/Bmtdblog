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

func IndexGet(c *gin.Context) {
	var (
		pageIndex int
		pageSize  = config.GetConfiguration().PageSize
		total     int
		page      string
		err       error
		posts     []*models.Post
		policy    *bluemonday.Policy
	)
	page = c.Query("page")
	pageIndex, _ = strconv.Atoi(page)
	if pageIndex <= 0 {
		pageIndex = 1
	}
	posts, err = models.ListPublishedPost("", pageIndex, pageSize)
	if err != nil {
		log.Error("models.ListPublishedPost error", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	total, err = models.CountPostByTag("")
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
		"user":            user,
		"pageIndex":       pageIndex,
		"totalPage":       int(math.Ceil(float64(total) / float64(pageSize))),
		"path":            c.Request.URL.Path,
		"maxReadPosts":    models.MustListMaxReadPost(),
		"maxCommentPosts": models.MustListMaxCommentPost(),
		"cfg":             config.GetConfiguration(),
	})
}
