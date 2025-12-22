package comment

import (
	"fmt"

	"github.com/dchest/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
)

func CommentPost(c *gin.Context) {
	var (
		err  error
		res  = gin.H{}
		post *models.Post
		cfg  = config.GetConfiguration()
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
	log.Debug("CommentPost", zap.Uint("userId", userId), zap.Uint("postId", pid), zap.String("content", content))
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
