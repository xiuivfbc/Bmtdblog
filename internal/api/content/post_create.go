package content

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

// @Summary 创建新文章
// @Description 创建新的博客文章并添加标签
// @Tags 内容管理
// @Accept x-www-form-urlencoded
// @Produce html
// @Param title formData string true "文章标题"
// @Param body formData string true "文章内容"
// @Param tags formData string true "文章标签ID，逗号分隔"
// @Param isPublished formData string false "是否发布(on表示发布)"
// @Success 301 {string} redirect "重定向到文章管理页面"
// @Failure 200 {string} html "创建失败的页面"
// @Router /admin/post/create [post]
func PostCreate(c *gin.Context) {
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"
	log.Debug("PostCreate", zap.String("title", title), zap.String("tags", tags), zap.Bool("isPublished", published))

	post := &models.Post{
		Title:       title,
		Body:        body,
		IsPublished: published,
	}
	err := post.Insert()
	if err != nil {
		c.HTML(http.StatusOK, "post/new.html", gin.H{
			"post":    post,
			"message": err.Error(),
			"user":    c.MustGet(common.ContextUserKey),
			"cfg":     config.GetConfiguration(),
		})
		return
	}

	// add tag for post
	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for _, tag := range tagArr {
			tagId, err := common.ParseUint(tag)
			if err != nil {
				continue
			}
			pt := &models.PostTag{
				PostId: post.ID,
				TagId:  tagId,
			}
			pt.Insert()
		}
	}
	c.Redirect(http.StatusMovedPermanently, "/admin/post")
}
