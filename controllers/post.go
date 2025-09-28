package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
)

func PostGet(c *gin.Context) {
	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostByIdWithCache(id)
	if err != nil || !post.IsPublished {
		Handle404(c)
		return
	}
	// 更新浏览数（异步，避免影响缓存和性能）
	go func() {
		post.View++
		post.UpdateView()
	}()

	user, _ := c.Get(ContextUserKey)
	c.HTML(http.StatusOK, "post/display.html", gin.H{
		"post": post,
		"user": user,
		"cfg":  system.GetConfiguration(),
	})
}

func PostNew(c *gin.Context) {
	c.HTML(http.StatusOK, "post/new.html", gin.H{
		"user": c.MustGet(ContextUserKey),
		"cfg":  system.GetConfiguration(),
	})
}

func PostCreate(c *gin.Context) {
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"

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
			"user":    c.MustGet(ContextUserKey),
			"cfg":     system.GetConfiguration(),
		})
		return
	}

	// add tag for post
	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for _, tag := range tagArr {
			tagId, err := parseUint(tag)
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

func PostEdit(c *gin.Context) {
	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
		return
	}
	post, err := models.GetPostByIdWithCache(id)
	if err != nil {
		Handle404(c)
		return
	}
	c.HTML(http.StatusOK, "post/modify.html", gin.H{
		"post": post,
		"user": c.MustGet(ContextUserKey),
		"cfg":  system.GetConfiguration(),
	})
}

func PostUpdate(c *gin.Context) {
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	body := c.PostForm("body")
	isPublished := c.PostForm("isPublished")
	published := isPublished == "on"

	id, err := ParamUint(c, "id")
	if err != nil {
		HandleMessage(c, err.Error())
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
			"user":    c.MustGet(ContextUserKey),
			"cfg":     system.GetConfiguration(),
		})
		return
	}
	// 删除tag
	models.DeletePostTagByPostId(post.ID)
	// 添加tag
	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for _, tag := range tagArr {
			tagId, err := parseUint(tag)
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

func PostPublish(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
	)
	defer writeJSON(c, res)
	id, err := ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post, err = models.GetPostByIdWithCache(id)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post.IsPublished = !post.IsPublished
	err = post.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func PostDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer writeJSON(c, res)
	id, err := ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post := &models.Post{}
	post.ID = id
	err = post.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	models.DeletePostTagByPostId(id)
	res["succeed"] = true
}

func PostIndex(c *gin.Context) {
	posts, _ := models.ListAllPost("")
	c.HTML(http.StatusOK, "admin/post.html", gin.H{
		"posts":    posts,
		"Active":   "posts",
		"user":     c.MustGet(ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      system.GetConfiguration(),
	})
}
