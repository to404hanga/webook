package repository

import (
	"context"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/article/domain"
	"webook/article/repository/dao"
)

//go:generate mockgen -source=./author.go -destination=./mocks/author.mock.go -package=repomocks AuthorRepository
type AuthorRepository interface {
	FindAuthor(ctx context.Context, id int64) (domain.Author, error)
}

type GrpcAuthorRepository struct {
	client userv1.UserServiceClient
	dao    dao.ArticleDAO
}

var _ AuthorRepository = (*GrpcAuthorRepository)(nil)

func NewGrpcAuthorRepository(client userv1.UserServiceClient, dao dao.ArticleDAO) AuthorRepository {
	return &GrpcAuthorRepository{
		client: client,
		dao:    dao,
	}
}

func (g *GrpcAuthorRepository) FindAuthor(ctx context.Context, id int64) (domain.Author, error) {
	article, err := g.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Author{}, err
	}
	user, err := g.client.Profile(ctx, &userv1.ProfileRequest{
		Id: article.AuthorId,
	})
	if err != nil {
		return domain.Author{}, err
	}
	return domain.Author{
		Id:   user.GetUser().GetId(),
		Name: user.GetUser().GetNickname(),
	}, nil
}
