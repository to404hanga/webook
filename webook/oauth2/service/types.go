package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"webook/oauth2/domain"

	"github.com/to404hanga/pkg404/logger"
)

//go:generate mockgen -source=./types.go -destination=./mocks/oauth2.mock.go -package=svcmocks Service
type Service interface {
	Auth2URL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appID     string
	appSecret string
	client    *http.Client
	logger    logger.Logger
}

const (
	authURLPattern        = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
	accessTokenURLPattern = `https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`
)

var (
	redirectURL = url.PathEscape(`https://meoying.com/oauth2/wechat/callback`)
)

func NewService(appId, appSecret string, logger logger.Logger) Service {
	return &service{
		appID:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
		logger:    logger,
	}
}

func (s *service) Auth2URL(ctx context.Context, state string) (string, error) {
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	accessTokenURL := fmt.Sprintf(accessTokenURLPattern, s.appID, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenURL, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("调用微信接口失败 errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		UnionID: res.UnionID,
		OpenID:  res.OpenID,
	}, nil
}

type Result struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`

	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}
