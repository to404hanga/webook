package grpc

import (
	"context"
	articlev1 "webook/api/proto/gen/article/v1"
	"webook/article/domain"
	"webook/article/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ArticleServiceServer struct {
	articlev1.UnimplementedArticleServiceServer
	svc service.ArticleService
}

func NewArticleServiceServer(svc service.ArticleService) *ArticleServiceServer {
	return &ArticleServiceServer{
		svc: svc,
	}
}

func (a *ArticleServiceServer) Register(server grpc.ServiceRegistrar) {
	articlev1.RegisterArticleServiceServer(server, a)
}

func (a *ArticleServiceServer) Save(ctx context.Context, req *articlev1.SaveRequest) (*articlev1.SaveResponse, error) {
	article := a.convertToDomain(req.GetArticle())
	id, err := a.svc.Save(ctx, article)
	return &articlev1.SaveResponse{
		Id: id,
	}, err
}

func (a *ArticleServiceServer) Publish(ctx context.Context, req *articlev1.PublishRequest) (*articlev1.PublishResponse, error) {
	article := a.convertToDomain(req.GetArticle())
	id, err := a.svc.Publish(ctx, article)
	return &articlev1.PublishResponse{
		Id: id,
	}, err
}

func (a *ArticleServiceServer) Withdraw(ctx context.Context, req *articlev1.WithdrawRequest) (*articlev1.WithdrawResponse, error) {
	err := a.svc.Withdraw(ctx, req.GetUid(), req.GetId())
	return &articlev1.WithdrawResponse{}, err
}

func (a *ArticleServiceServer) List(ctx context.Context, req *articlev1.ListRequest) (*articlev1.ListResponse, error) {
	articleList, err := a.svc.List(ctx, req.GetAuthor(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}
	list := transform.SliceFromSlice[domain.Article, *articlev1.Article](articleList, func(art domain.Article) *articlev1.Article {
		return a.convertToV(art)
	})
	return &articlev1.ListResponse{
		Articles: list,
	}, nil
}

func (a *ArticleServiceServer) GetById(ctx context.Context, req *articlev1.GetByIdRequest) (*articlev1.GetByIdResponse, error) {
	article, err := a.svc.GetById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &articlev1.GetByIdResponse{
		Article: a.convertToV(article),
	}, nil
}

func (a *ArticleServiceServer) GetPubById(ctx context.Context, req *articlev1.GetPubByIdRequest) (*articlev1.GetPubByIdResponse, error) {
	article, err := a.svc.GetPubById(ctx, req.GetId(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &articlev1.GetPubByIdResponse{
		Article: a.convertToV(article),
	}, nil
}

func (a *ArticleServiceServer) ListPub(ctx context.Context, req *articlev1.ListPubRequest) (*articlev1.ListPubResponse, error) {
	articleList, err := a.svc.ListPub(ctx, req.GetStartTime().AsTime(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}
	list := transform.SliceFromSlice[domain.Article, *articlev1.Article](articleList, func(art domain.Article) *articlev1.Article {
		return a.convertToV(art)
	})
	return &articlev1.ListPubResponse{
		Articles: list,
	}, nil
}

func (a *ArticleServiceServer) convertToV(article domain.Article) *articlev1.Article {
	art := articlev1.Article{}
	art.Id = article.Id
	art.Title = article.Title
	art.Content = article.Content
	art.Status = int32(article.Status)
	art.Author = &articlev1.Author{
		Id:   article.Author.Id,
		Name: article.Author.Name,
	}
	art.CreateTime = timestamppb.New(article.CreateTime)
	art.UpdateTime = timestamppb.New(article.UpdateTime)
	return &art
}

func (a *ArticleServiceServer) convertToDomain(article *articlev1.Article) domain.Article {
	art := domain.Article{}
	if article != nil {
		art.Id = article.GetId()
		art.Title = article.GetTitle()
		art.Content = article.GetContent()
		art.Status = domain.ArticleStatus(article.GetStatus())
		art.Author = domain.Author{
			Id:   article.GetAuthor().GetId(),
			Name: article.GetAuthor().GetName(),
		}
		art.CreateTime = article.GetCreateTime().AsTime()
		art.UpdateTime = article.GetUpdateTime().AsTime()
	}
	return art
}
