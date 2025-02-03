package grpc

import (
	"context"
	codev1 "webook/api/proto/gen/code/v1"
	"webook/code/service"

	"google.golang.org/grpc"
)

type CodeServiceServer struct {
	codev1.UnimplementedCodeServiceServer
	svc service.CodeService
}

func NewCodeServiceServer(svc service.CodeService) *CodeServiceServer {
	return &CodeServiceServer{svc: svc}
}

func (s *CodeServiceServer) Register(server grpc.ServiceRegistrar) {
	codev1.RegisterCodeServiceServer(server, s)
}

func (s *CodeServiceServer) Send(ctx context.Context, req *codev1.CodeSendRequest) (*codev1.CodeSendResponse, error) {
	err := s.svc.Send(ctx, req.GetBiz(), req.GetPhone())
	return &codev1.CodeSendResponse{}, err
}

func (s *CodeServiceServer) Verify(ctx context.Context, req *codev1.VerifyRequest) (*codev1.VerifyResponse, error) {
	ok, err := s.svc.Verify(ctx, req.GetBiz(), req.GetPhone(), req.GetInputCode())
	return &codev1.VerifyResponse{Success: ok}, err
}
