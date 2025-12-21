package content

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
)

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
