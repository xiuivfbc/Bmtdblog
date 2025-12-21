package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func PageGet(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	page, err := models.GetPageById(id)
	if err != nil || !page.IsPublished {
		common.Handle404(c)
		return
	}
	page.View++
	page.UpdateView()
	user, _ := c.Get(common.ContextUserKey)
	c.HTML(http.StatusOK, "page/display.html", gin.H{
		"page": page,
		"user": user,
		"cfg":  config.GetConfiguration(),
	})
}
