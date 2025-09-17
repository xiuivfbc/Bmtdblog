package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cihub/seelog"

	"github.com/denisbakhtin/sitemap"
	"github.com/gin-gonic/gin"
	"github.com/wangsongyan/wblog/helpers"
	"github.com/wangsongyan/wblog/models"
	"github.com/wangsongyan/wblog/system"
)

const (
	SessionKey         = "UserID"       // session key
	ContextUserKey     = "User"         // context user key
	SessionGithubState = "GITHUB_STATE" // GitHub state session key
	SessionCaptcha     = "GIN_CAPTCHA"  // captcha session key
)

func Handle404(c *gin.Context) {
	HandleMessage(c, "Sorry,I lost myself!")
}

func HandleMessage(c *gin.Context, message string) {
	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusNotFound, "errors/error.html", gin.H{
		"message": message,
		"user":    user,
		"cfg":     system.GetConfiguration(),
	})
}

func sendMail(to, subject, body string) error {
	cfg := system.GetConfiguration()
	if !cfg.Smtp.Enabled {
		return nil
	}
	return helpers.SendToMail(cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Host, to, subject, body, "html")
}

func NotifyEmail(subject, body string) error {
	notifyEmailsStr := system.GetConfiguration().NotifyEmails
	if notifyEmailsStr != "" {
		notifyEmails := strings.Split(notifyEmailsStr, ";")
		emails := make([]string, 0)
		for _, email := range notifyEmails {
			if email != "" {
				emails = append(emails, email)
			}
		}
		if len(emails) > 0 {
			return sendMail(strings.Join(emails, ";"), subject, body)
		}
	}
	return nil
}

func CreateXMLSitemap() (err error) {
	cfg := system.GetConfiguration()
	folder := path.Join(helpers.GetCurrentDirectory(), cfg.PublicDir, "sitemap")
	err = os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		seelog.Errorf("create folder:%v", err)
		return
	}
	domain := cfg.Domain
	now := helpers.GetCurrentTime()
	items := make([]sitemap.Item, 0)

	items = append(items, sitemap.Item{
		Loc:        domain,
		LastMod:    now,
		Changefreq: "daily",
		Priority:   1,
	})

	posts, err := models.ListPublishedPost("", 0, 0)
	if err != nil {
		seelog.Errorf("models.ListPublishedPost:%v", err)
		return
	}
	for _, post := range posts {
		items = append(items, sitemap.Item{
			Loc:        fmt.Sprintf("%s/post/%d", domain, post.ID),
			LastMod:    post.UpdatedAt,
			Changefreq: "weekly",
			Priority:   0.9,
		})
	}

	pages, err := models.ListPublishedPage()
	if err != nil {
		seelog.Errorf("models.ListPublishedPage:%v", err)
		return
	}
	for _, page := range pages {
		items = append(items, sitemap.Item{
			Loc:        fmt.Sprintf("%s/page/%d", domain, page.ID),
			LastMod:    page.UpdatedAt,
			Changefreq: "monthly",
			Priority:   0.8,
		})
	}

	err = sitemap.SiteMap(path.Join(folder, "sitemap1.xml.gz"), items)
	if err != nil {
		seelog.Errorf("sitemap.SiteMap:%v", err)
		return
	}
	err = sitemap.SiteMapIndex(folder, "sitemap_index.xml", domain+"/static/sitemap/")
	if err != nil {
		seelog.Errorf("sitemap.SiteMapIndex:%v", err)
		return
	}
	return
}

func QueryUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.Query(key))
}

func ParamUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.Param(key))
}

func PostFormUint(c *gin.Context, key string) (uint, error) {
	return parseUint(c.PostForm(key))
}

func parseUint(value string) (uint, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

func writeJSON(ctx *gin.Context, h gin.H) {
	if _, ok := h["succeed"]; !ok {
		h["succeed"] = false
	}
	ctx.JSON(http.StatusOK, h)
}
