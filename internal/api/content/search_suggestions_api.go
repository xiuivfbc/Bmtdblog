package content

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary 获取搜索建议
// @Description 根据输入前缀获取搜索建议列表
// @Tags 搜索功能
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词前缀(至少2个字符)"
// @Success 200 {object} map[string]interface{} "{"suggestions":["建议1","建议2"]}"
// @Failure 200 {object} map[string]interface{} "{"suggestions":[]}"
// @Router /api/search/suggestions [get]
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
