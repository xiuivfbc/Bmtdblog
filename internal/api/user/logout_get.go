package user

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
)

func LogoutGet(c *gin.Context) {
	log.Debug("LogoutGet")
	s := sessions.Default(c)
	s.Clear()
	s.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
