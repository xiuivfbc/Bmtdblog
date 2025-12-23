package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func PostEdit(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostByIdWithCache(id)
	if err != nil {
		common.Handle404(c)
		return
	}
	c.HTML(http.StatusOK, "post/modify.html", gin.H{
		"post": post,
		"user": c.MustGet(common.ContextUserKey),
		"cfg":  config.GetConfiguration(),
	})
}
