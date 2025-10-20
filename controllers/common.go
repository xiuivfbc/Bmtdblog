package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/denisbakhtin/sitemap"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/helpers"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
	"golang.org/x/oauth2"
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

	// 尝试使用邮件队列异步发送
	if err := system.PushEmailTask(to, subject, body); err != nil {
		// 队列失败时降级到同步发送
		system.Logger.Warn("邮件队列发送失败，降级到同步发送", "err", err)
		return helpers.SendToMail(cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Host, to, subject, body, "html")
	}

	return nil
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
		system.Logger.Error("create folder error", "err", err)
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
		system.Logger.Error("models.ListPublishedPost error", "err", err)
		return
	}
	for _, post := range posts {
		items = append(items, sitemap.Item{
			Loc:        fmt.Sprintf("%s/post/%d", domain, post.ID),
			LastMod:    *post.UpdatedAt,
			Changefreq: "weekly",
			Priority:   0.9,
		})
	}

	pages, err := models.ListPublishedPage()
	if err != nil {
		system.Logger.Error("models.ListPublishedPage error", "err", err)
		return
	}
	for _, page := range pages {
		items = append(items, sitemap.Item{
			Loc:        fmt.Sprintf("%s/page/%d", domain, page.ID),
			LastMod:    *page.UpdatedAt,
			Changefreq: "monthly",
			Priority:   0.8,
		})
	}

	err = sitemap.SiteMap(path.Join(folder, "sitemap1.xml.gz"), items)
	if err != nil {
		system.Logger.Error("sitemap.SiteMap error", "err", err)
		return
	}
	err = sitemap.SiteMapIndex(folder, "sitemap_index.xml", domain+"/static/sitemap/")
	if err != nil {
		system.Logger.Error("sitemap.SiteMapIndex error", "err", err)
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

func saveToken(path string, token *oauth2.Token) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var token oauth2.Token
	err = json.NewDecoder(f).Decode(&token)
	return &token, err
}
