package interaction

import (
	"fmt"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"github.com/xiuivfbc/bmtdblog/internal/system"
)

func CommentPost(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
		cfg  = system.GetConfiguration()
	)
	defer common.WriteJSON(c, res)
	s := sessions.Default(c)
	userId := s.Get(common.SessionKey).(uint)
	verifyCode := c.PostForm("verifyCode")
	captchaId := s.Get(common.SessionCaptcha).(string)
	s.Delete(common.SessionCaptcha)
	if !captcha.VerifyString(captchaId, verifyCode) {
		res["message"] = "error verifyCode"
		return
	}

	content := c.PostForm("content")
	if len(content) == 0 {
		res["message"] = "content cannot be empty."
		return
	}
	pid, err := common.ParamUint(c, "postId")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	post, err = models.GetPostByIdWithCache(pid)
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		PostID:  pid,
		Content: content,
		UserID:  userId,
	}
	err = comment.Insert()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	common.NotifyEmail(fmt.Sprintf("[%s]您有一条新评论", cfg.Title), fmt.Sprintf("<a href=\"%s/post/%d\" target=\"_blank\">%s</a>:%s", cfg.Domain, post.ID, post.Title, content))
	res["succeed"] = true
}

func CommentDelete(c *gin.Context) {
	var (
		err error
		res = gin.H{}
		cid uint
	)
	defer common.WriteJSON(c, res)

	s := sessions.Default(c)
	userId := s.Get(common.SessionKey).(uint)

	cid, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := &models.Comment{
		UserID: userId,
	}
	comment.ID = cid
	err = comment.Delete()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func CommentRead(c *gin.Context) {
	var (
		id  uint
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	id, err = common.ParamUint(c, "id")
	if err != nil {
		res["message"] = err.Error()
		return
	}
	comment := new(models.Comment)
	comment.ID = id
	err = comment.Update()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}

func CommentReadAll(c *gin.Context) {
	var (
		err error
		res = gin.H{}
	)
	defer common.WriteJSON(c, res)
	err = models.SetAllCommentRead()
	if err != nil {
		res["message"] = err.Error()
		return
	}
	res["succeed"] = true
}
