package user

import (
	"context"

	"github.com/xiuivfbc/bmtdblog/internal/common"
	"github.com/xiuivfbc/bmtdblog/internal/common/log"
	"github.com/xiuivfbc/bmtdblog/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func exchangeTokenByCode(code string) (accessToken string, err error) {
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
