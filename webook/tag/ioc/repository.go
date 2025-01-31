package ioc

import (
	"context"
	"time"
	"webook/tag/repository"
	"webook/tag/repository/cache"
	"webook/tag/repository/dao"

	"github.com/to404hanga/pkg404/logger"
)

func InitRepository(dao dao.TagDAO, cache cache.TagCache, l logger.Logger) repository.TagRepository {
	repo := repository.NewCachedTagRepository(dao, cache, l)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		repo.PreloadUserTags(ctx)
	}()

	return repo
}
