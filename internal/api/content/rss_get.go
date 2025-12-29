package content

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

// @Summary RSS订阅
// @Description 获取博客RSS订阅内容，XML格式
// @Tags 内容浏览
// @Accept */*
// @Produce application/xml
// @Success 200 {string} string "RSS订阅内容（XML格式）"
// @Failure 500 {string} string "服务器内部错误"
// @Router /rss [get]
func RssGet(c *gin.Context) {
	cfg := config.GetConfiguration()
	now := common.GetCurrentTime()
	domain := config.GetConfiguration().Domain
	feed := &feeds.Feed{
		Title:       cfg.Title,
		Link:        &feeds.Link{Href: domain},
		Description: cfg.Seo.Description,
		Author:      &feeds.Author{Name: cfg.Seo.Author.Name, Email: cfg.Seo.Author.Email},
		Created:     now,
	}

	feed.Items = make([]*feeds.Item, 0)
	posts, err := models.ListPublishedPost("", 0, 0)
	if err != nil {
		log.Error("models.ListPublishedPost err", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	for _, post := range posts {
		item := &feeds.Item{
			Id:          fmt.Sprintf("%s/post/%d", domain, post.ID),
			Title:       post.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/post/%d", domain, post.ID)},
			Description: string(post.Excerpt()),
			Created:     now,
		}
		feed.Items = append(feed.Items, item)
	}
	rss, err := feed.ToRss()
	if err != nil {
		log.Error("feed.ToRss err", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteString(rss)
}
