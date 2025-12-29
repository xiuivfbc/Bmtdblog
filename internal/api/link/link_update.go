package link

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 更新链接
// @Description 更新已存在的友情链接信息
// @Tags 链接管理
// @Accept json
// @Produce json
// @Param id path int true "链接ID"
// @Param name body string true "链接名称"
// @Param url body string true "链接URL"
// @Param sort body int false "排序值"
// @Success 200 {object} map[string]interface{} "{"succeed":true,"link":{"id":1,"name":"链接名","url":"https://example.com","sort":1}}"
// @Failure 200 {object} map[string]interface{} "{"succeed":false,"message":"错误信息"}"
// @Router /admin/link/{id} [post]
func LinkUpdate(c *gin.Context) {
	var (
		link models.Link
		res  = gin.H{}
		err  error
		id   int
	)
	defer common.WriteJSON(c, res)
	// 获取ID
	idStr := c.Param("id")
	id, err = strconv.Atoi(idStr)
	if err != nil {
		res["message"] = "Invalid ID"
		return
	}
	// 绑定表单数据
	if err = c.ShouldBind(&link); err != nil {
		res["message"] = err.Error()
		return
	}
	// 设置ID
	link.ID = uint(id)
	log.Debug("LinkUpdate", zap.Uint("id", link.ID), zap.String("name", link.Name), zap.String("url", link.Url), zap.Int("sort", link.Sort))
	// 验证数据
	if link.Name == "" || link.Url == "" {
		res["message"] = "Name and URL are required"
		return
	}
	// 更新链接
	if err = link.Update(); err != nil {
		res["message"] = err.Error()
		return
	}
	res["link"] = link
	res["succeed"] = true
}
