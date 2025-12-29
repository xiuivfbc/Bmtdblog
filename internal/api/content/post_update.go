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

// @Summary 更新文章
// @Description 更新已存在的博客文章，包括内容、标题、标签和发布状态
// @Tags 内容管理
// @Accept x-www-form-urlencoded
// @Produce html
// @Param id path uint true "文章ID"
// @Param title formData string true "文章标题"
// @Param body formData string true "文章内容"
// @Param tags formData string true "文章标签ID，逗号分隔"
// @Param isPublished formData string false "是否发布(on表示发布)"
// @Success 301 {string} redirect "重定向到文章管理页面"
// @Failure 200 {string} html "更新失败的页面"
// @Router /admin/post/{id} [post]
func PostUpdate(c *gin.Context) {
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"

	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	log.Debug("PostUpdate", zap.Uint("id", id), zap.String("title", title), zap.String("tags", tags), zap.Bool("isPublished", published))

	post := &models.Post{
		Title:       title,
		Body:        body,
		IsPublished: published,
	}
	post.ID = id
	err = post.Update()
	if err != nil {
		c.HTML(http.StatusOK, "post/modify.html", gin.H{
			"post":    post,
			"message": err.Error(),
			"user":    c.MustGet(common.ContextUserKey),
			"cfg":     config.GetConfiguration(),
		})
		return
	}
	// 删除tag
	models.DeletePostTagByPostId(post.ID)
	// 添加tag
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
