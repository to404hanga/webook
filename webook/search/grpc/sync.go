package grpc

import (
	"context"
	searchv1 "webook/api/proto/gen/search/v1"
	"webook/search/domain"
	"webook/search/service"

	"google.golang.org/grpc"
)

type SyncServiceServer struct {
	searchv1.UnimplementedSyncServiceServer
	syncSvc service.SyncService
}

func NewSyncServiceServer(syncSvc service.SyncService) *SyncServiceServer {
	return &SyncServiceServer{syncSvc: syncSvc}
}

func (s *SyncServiceServer) Register(server grpc.ServiceRegistrar) {
	searchv1.RegisterSyncServiceServer(server, s)
}

func (s *SyncServiceServer) InputUser(ctx context.Context, req *searchv1.InputUserRequest) (*searchv1.InputUserResponse, error) {
	err := s.syncSvc.InputUser(ctx, s.toDomainUser(req.GetUser()))
	return &searchv1.InputUserResponse{}, err
}

func (s *SyncServiceServer) InputArticle(ctx context.Context, req *searchv1.InputArticleRequest) (*searchv1.InputArticleResponse, error) {
	err := s.syncSvc.InputArticle(ctx, s.toDomainArticle(req.GetArticle()))
	return &searchv1.InputArticleResponse{}, err
}

func (s *SyncServiceServer) InputAny(ctx context.Context, req *searchv1.InputAnyRequest) (*searchv1.InputAnyResponse, error) {
	err := s.syncSvc.InputAny(ctx, req.GetIndexName(), req.GetDocId(), req.GetData())
	return &searchv1.InputAnyResponse{}, err
}

func (s *SyncServiceServer) toDomainUser(user *searchv1.User) domain.User {
	return domain.User{
		Id:       user.GetId(),
		Nickname: user.GetNickname(),
		Email:    user.GetEmail(),
	}
}

func (s *SyncServiceServer) toDomainArticle(article *searchv1.Article) domain.Article {
	return domain.Article{
		Id:      article.GetId(),
		Title:   article.GetTitle(),
		Content: article.GetContent(),
		Status:  article.GetStatus(),
		Tags:    article.GetTags(),
	}
}
