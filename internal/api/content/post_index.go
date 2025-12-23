package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func PostIndex(c *gin.Context) {
	posts, _ := models.ListAllPost("")
	c.HTML(http.StatusOK, "admin/post.html", gin.H{
		"posts":    posts,
		"Active":   "posts",
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      config.GetConfiguration(),
	})
}
