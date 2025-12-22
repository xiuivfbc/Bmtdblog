package link

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

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
