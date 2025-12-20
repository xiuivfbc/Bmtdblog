package system

import (
	"net/http"
	"strconv"

	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"

	"github.com/gin-gonic/gin"
)

func LinkIndex(c *gin.Context) {
	links, _ := models.ListLinks()
	c.HTML(http.StatusOK, "admin/link.html", gin.H{
		"links":    links,
		"user":     c.MustGet(common.ContextUserKey),
		"comments": models.MustListUnreadComment(),
		"cfg":      system.GetConfiguration(),
	})
}

func LinkCreate(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		sort int
	)
	defer common.WriteJSON(c, res)
	name := c.PostForm("name")
	url := c.PostForm("url")
	if len(name) == 0 || len(url) == 0 {
		res["message"] = "error parameter"
		return
	}
	sort, _ = strconv.Atoi(c.PostForm("sort"))
	link := &models.Link{
		Name: name,
		Url:  url,
		Sort: sort,
	}
	err = link.Insert()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func LinkUpdate(c *gin.Context) {
	var (
		id   uint
		sort int
		err  error
		res  = gin.H{}
	)
	defer common.WriteJSON(c, res)
	name := c.PostForm("name")
	url := c.PostForm("url")
	if len(name) == 0 || len(url) == 0 {
		res["message"] = "error parameter"
		return
	}
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	sort, _ = strconv.Atoi(c.PostForm("sort"))
	link := &models.Link{
		Name: name,
		Url:  url,
		Sort: sort,
	}
	link.ID = id
	err = link.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func LinkGet(c *gin.Context) {
	id, err := common.ParamUint(c, "id")
	if err != nil {
		common.HandleMessage(c, err.Error())
		return
	}
	link, err := models.GetLinkById(id)
	if err != nil {
		system.Logger.Error("models.GetLinkById error", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	link.View++
	link.Update()
	c.Redirect(http.StatusFound, link.Url)
}

func LinkDelete(c *gin.Context) {
	var (
		err error
		id  uint
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}

	link := new(models.Link)
	link.ID = id
	err = link.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
