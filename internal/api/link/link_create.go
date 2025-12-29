package link

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 创建链接
// @Description 创建新的友情链接
// @Tags 链接管理
// @Accept x-www-form-urlencoded
// @Produce json
// @Param name formData string true "链接名称"
// @Param url formData string true "链接URL"
// @Param sort formData int false "排序值"
// @Success 200 {object} map[string]interface{} "{"succeed":true}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/link/create [post]
func LinkCreate(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		sort int
	)
	defer common.WriteJSON(c, res)
	name := c.PostForm("name")
	url := c.PostForm("url")
	if len(name) == 0 || len(url) == 0 {
		res["message"] = "error parameter"
		return
	}
	sort, _ = strconv.Atoi(c.PostForm("sort"))
	log.Debug("LinkCreate", zap.String("name", name), zap.String("url", url), zap.Int("sort", sort))
	link := &models.Link{
		Name: name,
		Url:  url,
		Sort: sort,
	}
	err = link.Insert()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
