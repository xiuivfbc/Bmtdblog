package content

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

func PageCreate(c *gin.Context) {
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"

	page := &models.Page{
		Title:       title,
		Body:        body,
		IsPublished: published,
	}
	err := page.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "page/new.html", gin.H{
			"message": err.Error(),
			"page":    page,
			"user":    c.MustGet(common.ContextUserKey),
			"cfg":     config.GetConfiguration(),
		})
		return
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/page")
}
