package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xiuivfbc/bmtdblog/helpers"
	"github.com/xiuivfbc/bmtdblog/models"
	"github.com/xiuivfbc/bmtdblog/system"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func AuthGet(c *gin.Context) {
	authType := c.Param("authType")

	session := sessions.Default(c)
	uuid := helpers.UUID()
	session.Delete(SessionGithubState)
	session.Set(SessionGithubState, uuid)
	session.Save()

	cfg := system.GetConfiguration()

	authurl := "/signin"
	switch authType {
	case "github":
		authurl = fmt.Sprintf(cfg.Github.AuthUrl, cfg.Github.ClientId, uuid)
	default:
	}
	c.Redirect(http.StatusFound, authurl)
}

func Oauth2Callback(c *gin.Context) {
	var (
		userInfo *GithubUserInfo
		user     *models.User
	)
	code := c.Query("code")
	state := c.Query("state")

	session := sessions.Default(c)
	if len(state) == 0 || state != session.Get(SessionGithubState) {
		c.Abort()
		return
	}
	session.Delete(SessionGithubState)
	session.Save()
	token, err := exchangeTokenByCode(code)
	if err != nil {
		system.Logger.Error("exchangeTokenByCode error", "err", err)
		c.Redirect(http.StatusMovedPermanently, "/signin")
		return
	}

	userInfo, err = getGithubUserInfoByAccessToken(token)
	fmt.Println(userInfo)
	if err != nil {
		system.Logger.Error("getGithubUserInfoByAccessToken error", "err", err)
		c.Redirect(http.StatusMovedPermanently, "/signin")
		return
	}

	sessionUser, exists := c.Get(ContextUserKey)
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
				HandleMessage(c, err.Error())
				return
			}
		}
	}

	if err == nil {
		s := sessions.Default(c)
		s.Clear()
		s.Set(SessionKey, user.ID)
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
	var (
		token *oauth2.Token
		cfg   = system.GetConfiguration()
	)
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Github.ClientId,
		ClientSecret: cfg.Github.ClientSecret,
		RedirectURL:  cfg.Github.RedirectURL,
		Endpoint:     github.Endpoint,
	}
	token, err = oauthCfg.Exchange(context.Background(), code)
	if err != nil {
		return
	}
	accessToken = token.AccessToken
	if err := saveToken("./request.token", token); err != nil {
		system.Logger.Error("saveToken error", "err", err)
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
