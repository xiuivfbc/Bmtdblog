package user

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"github.com/xiuivfbc/bmtdblog/internal/models"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// @Summary GitHub OAuth回调
// @Description 处理GitHub OAuth2认证回调，完成用户登录或绑定
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param code query string true "GitHub授权码"
// @Param state query string true "状态参数，用于防止CSRF攻击"
// @Success 301 {string} redirect "登录成功后重定向到首页或管理后台"
// @Failure 301 {string} redirect "认证失败后重定向到登录页"
// @Router /oauth2callback [get]
func Oauth2Callback(c *gin.Context) {
	var (
		userInfo *GithubUserInfo
		user     *models.User
	)
	code := c.Query("code")
	state := c.Query("state")
	log.Debug("Oauth2Callback", zap.String("code", code), zap.String("state", state))

	session := sessions.Default(c)
	if len(state) == 0 || state != session.Get(common.SessionGithubState) {
		c.Abort()
		return
	}
	session.Delete(common.SessionGithubState)
	session.Save()
	token, err := exchangeTokenByCode(code)
	if err != nil {
		log.Error("exchangeTokenByCode error", "err", err)
		c.Redirect(http.StatusMovedPermanently, "/signin")
		return
	}

	userInfo, err = getGithubUserInfoByAccessToken(token)
	fmt.Println(userInfo)
	if err != nil {
		log.Error("getGithubUserInfoByAccessToken error", "err", err)
		c.Redirect(http.StatusMovedPermanently, "/signin")
		return
	}

	sessionUser, exists := c.Get(common.ContextUserKey)
	if exists { // 已登录
		user = sessionUser.(*models.User)
		if _, e := models.IsGithubIdExists(userInfo.Login, user.ID); e != nil { // 未绑定
			if user.IsAdmin {
				user.GithubLoginId = userInfo.Login
			}
			user.AvatarUrl = userInfo.AvatarURL
			user.GithubUrl = userInfo.HTMLURL
			err = user.UpdateGithubUserInfo()
		} else {
			err = errors.New("this github loginId has bound another account.")
		}
	} else {
		user = &models.User{
			GithubLoginId: userInfo.Login,
			AvatarUrl:     userInfo.AvatarURL,
			GithubUrl:     userInfo.HTMLURL,
		}
		user, err = user.FirstOrCreate()
		if err == nil {
			if user.LockState {
				err = errors.New("Your account have been locked.")
				common.HandleMessage(c, err.Error())
				return
			}
		}
	}

	if err == nil {
		s := sessions.Default(c)
		s.Clear()
		s.Set(common.SessionKey, user.ID)
		s.Save()
		if user.IsAdmin {
			c.Redirect(http.StatusMovedPermanently, "/admin/index")
		} else {
			c.Redirect(http.StatusMovedPermanently, "/")
		}
		return
	}
}

func exchangeTokenByCode(code string) (accessToken string, err error) {
	log.Debug("exchangeTokenByCode", zap.String("code", code))
	var (
		token *oauth2.Token
		cfg   = config.GetConfiguration()
	)
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Github.ClientId,
		ClientSecret: cfg.Github.ClientSecret,
		RedirectURL:  cfg.Github.RedirectUrl,
		Endpoint:     github.Endpoint,
	}
	token, err = oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		return
	}
	accessToken = token.AccessToken
	if err := common.SaveToken("./request.token", token); err != nil {
		log.Error("saveToken error", "err", err)
	}
	return
}


func getGithubUserInfoByAccessToken(token string) (*GithubUserInfo, error) {
	var (
		resp *http.Response
		req  *http.Request
		body []byte
		err  error
	)
	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var userInfo GithubUserInfo
	err = json.Unmarshal(body, &userInfo)
	return &userInfo, err
}

type GithubUserInfo struct {
	AvatarURL         string      `json:"avatar_url"`
	Bio               interface{} `json:"bio"`
	Blog              string      `json:"blog"`
	Company           interface{} `json:"company"`
	CreatedAt         string      `json:"created_at"`
	Email             interface{} `json:"email"`
	EventsURL         string      `json:"events_url"`
	Followers         int         `json:"followers"`
	FollowersURL      string      `json:"followers_url"`
	Following         int         `json:"following"`
	FollowingURL      string      `json:"following_url"`
	GistsURL          string      `json:"gists_url"`
	GravatarID        string      `json:"gravatar_id"`
	Hireable          interface{} `json:"hireable"`
	HTMLURL           string      `json:"html_url"`
	ID                int         `json:"id"`
	Location          interface{} `json:"location"`
	Login             string      `json:"login"`
	Name              interface{} `json:"name"`
	OrganizationsURL  string      `json:"organizations_url"`
	PublicGists       int         `json:"public_gists"`
	PublicRepos       int         `json:"public_repos"`
	ReceivedEventsURL string      `json:"received_events_url"`
	ReposURL          string      `json:"repos_url"`
	SiteAdmin         bool        `json:"site_admin"`
	StarredURL        string      `json:"starred_url"`
	SubscriptionsURL  string      `json:"subscriptions_url"`
	Type              string      `json:"type"`
	UpdatedAt         string      `json:"updated_at"`
	URL               string      `json:"url"`
}
