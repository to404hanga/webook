package grpc

import (
	"context"
	searchv1 "webook/api/proto/gen/search/v1"
	"webook/search/domain"
	"webook/search/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
)

type SearchServiceServer struct {
	searchv1.UnimplementedSearchServiceServer
	svc service.SearchService
}

func NewSearchServiceServer(svc service.SearchService) *SearchServiceServer {
	return &SearchServiceServer{svc: svc}
}

func (s *SearchServiceServer) Register(server grpc.ServiceRegistrar) {
	searchv1.RegisterSearchServiceServer(server, s)
}

func (s *SearchServiceServer) Search(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	resp, err := s.svc.Search(ctx, req.GetUid(), req.GetExpression())
	if err != nil {
		return nil, err
	}
	return &searchv1.SearchResponse{
		User: &searchv1.UserResult{
			Users: transform.SliceFromSlice[domain.User, *searchv1.User](resp.Users, func(u domain.User) *searchv1.User {
				return &searchv1.User{
					Id:       u.Id,
					Email:    u.Email,
					Nickname: u.Nickname,
					Phone:    u.Phone,
				}
			}),
		},
		Article: &searchv1.ArticleResult{
			Articles: transform.SliceFromSlice[domain.Article, *searchv1.Article](resp.Articles, func(a domain.Article) *searchv1.Article {
				return &searchv1.Article{
					Id:      a.Id,
					Title:   a.Title,
					Status:  a.Status,
					Content: a.Content,
				}
			}),
		},
	}, nil
}
