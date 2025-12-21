package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/config"
)

func SigninGet(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/signin.html", gin.H{
		"cfg": config.GetConfiguration(),
	})
}
