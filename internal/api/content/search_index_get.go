package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// SearchIndexGet 搜索首页
func SearchIndexGet(c *gin.Context) {
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "search/index.html", gin.H{
		"user":    user,
		"allTags": models.MustListTag(),
		"cfg":     config.GetConfiguration(),
	})
}
