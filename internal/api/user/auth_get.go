package user

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func AuthGet(c *gin.Context) {
	authType := c.Param("authType")

	session := sessions.Default(c)
	uuid := common.UUID()
	session.Delete(common.SessionGithubState)
	session.Set(common.SessionGithubState, uuid)
	session.Save()

	cfg := system.GetConfiguration()

	authurl := "/signin"
	switch authType {
	case "github":
		authurl = fmt.Sprintf(cfg.Github.AuthUrl, cfg.Github.ClientId, uuid)
	default:
	}
	c.Redirect(http.StatusFound, authurl)
}
