package service

import (
	"context"
	"fmt"
	"net/http"
	"webook/im/domain"

	"github.com/ecodeclub/ekit/net/httpx"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -source=./user.go -destination=./mocks/user.mock.go -package=svcmocks UserService
type UserService interface {
	Sync(ctx context.Context, user domain.User) error
}

type RestUserService struct {
	base   string // HTTP 请求的域名端口
	secret string // 默认是 openIM123
	client *http.Client
}

var _ UserService = (*RestUserService)(nil)

func NewRestUserService(base, secret string) UserService {
	return &RestUserService{
		base:   base,
		secret: secret,
		client: http.DefaultClient,
	}
}

func (r *RestUserService) Sync(ctx context.Context, user domain.User) error {
	var operationId string
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		operationId = spanCtx.TraceID().String()
	} else {
		operationId = uuid.New().String()
	}
	var resp response
	// TODO 替换为自己实现的 httpx 版本
	err := httpx.NewRequest(ctx, http.MethodPost, r.base+"/user/user_register").AddHeader("operationID", operationId).JSONBody(request{
		Secret: r.secret,
		Users:  []domain.User{user},
	}).Client(r.client).Do().JSONScan(&resp)
	if err != nil {
		return err
	}
	if resp.ErrCode != 0 {
		return fmt.Errorf("同步用户数据失败 %v", resp)
	}
	return nil
}

type request struct {
	Secret string        `json:"secret"`
	Users  []domain.User `json:"users"`
}

type response struct {
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
	ErrDlt  string `json:"errDlt"`
}
