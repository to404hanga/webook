package grpc

import (
	"context"
	oauth2v1 "webook/api/proto/gen/oauth2/v1"
	"webook/oauth2/service"

	"google.golang.org/grpc"
)

type Oauth2ServiceServer struct {
	oauth2v1.UnimplementedOauth2ServiceServer
	svc service.Service
}

func NewOauth2ServiceServer(svc service.Service) *Oauth2ServiceServer {
	return &Oauth2ServiceServer{svc: svc}
}

func (o *Oauth2ServiceServer) Register(server grpc.ServiceRegistrar) {
	oauth2v1.RegisterOauth2ServiceServer(server, o)
}

func (o *Oauth2ServiceServer) AuthURL(ctx context.Context, req *oauth2v1.AuthURLRequest) (*oauth2v1.AuthURLResponse, error) {
	url, err := o.svc.Auth2URL(ctx, req.GetState())
	return &oauth2v1.AuthURLResponse{
		Url: url,
	}, err
}

func (o *Oauth2ServiceServer) VerifyCode(ctx context.Context, req *oauth2v1.VerifyCodeRequest) (*oauth2v1.VerifyCodeResponse, error) {
	info, err := o.svc.VerifyCode(ctx, req.GetCode())
	return &oauth2v1.VerifyCodeResponse{
		OpenId:  info.OpenID,
		UnionId: info.UnionID,
	}, err
}
