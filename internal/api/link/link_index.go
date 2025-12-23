package link

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func LinkIndex(c *gin.Context) {
	links, _ := models.ListLinks()
	c.HTML(http.StatusOK, "admin/link.html", gin.H{
		"links":    links,
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      config.GetConfiguration(),
	})
}
