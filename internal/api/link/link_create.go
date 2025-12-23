package link

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

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
