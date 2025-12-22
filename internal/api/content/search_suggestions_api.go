package content

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// SearchSuggestionsAPI 搜索建议API
func SearchSuggestionsAPI(c *gin.Context) {
	prefix := strings.TrimSpace(c.Query("q"))
	if len(prefix) < 2 {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	suggestions, err := models.GetSearchSuggestions(prefix, 10)
	if err != nil {
		log.Error("获取搜索建议失败", "error", err, "prefix", prefix)
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}
