package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func AdminIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"pageCount":    models.CountPage(),
		"postCount":    models.CountPost(),
		"tagCount":     models.CountTag(),
		"commentCount": models.CountComment(),
		"user":         c.MustGet(common.ContextUserKey),
		"comments":     models.MustListUnreadComment(),
		"cfg":          config.GetConfiguration(),
	})
}
