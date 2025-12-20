package content

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func RssGet(c *gin.Context) {
	cfg := system.GetConfiguration()
	now := common.GetCurrentTime()
	domain := system.GetConfiguration().Domain
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
		system.Logger.Error("models.ListPublishedPost err", "err", err)
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
		system.Logger.Error("feed.ToRss err", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Writer.WriteString(rss)
}
