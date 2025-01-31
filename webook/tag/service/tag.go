package service

import (
	"context"
	"time"
	"webook/tag/domain"
	"webook/tag/events"
	"webook/tag/repository"

	"github.com/to404hanga/pkg404/logger"
	"github.com/to404hanga/pkg404/stl/transform"
)

type tagService struct {
	repo     repository.TagRepository
	l        logger.Logger
	producer events.Producer
}

var _ TagService = (*tagService)(nil)

func NewTagService(repo repository.TagRepository, l logger.Logger, producer events.Producer) TagService {
	return &tagService{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}

func (t *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return t.repo.GetTags(ctx, uid)
}

func (t *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return t.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}

func (t *tagService) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	return t.repo.GetBizTags(ctx, uid, biz, bizId)
}

func (t *tagService) AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tagIds []int64) error {
	err := t.repo.BindTagToBiz(ctx, uid, biz, bizId, tagIds)
	if err != nil {
		t.l.Error("贴标签失败", logger.Error(err), logger.Int64("uid", uid), logger.String("biz", biz), logger.Int64("biz_id", bizId), logger.Slice[int64]("tag_ids", tagIds))
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		tags, err := t.repo.GetTagsById(ctx, tagIds)
		cancel()
		if err != nil {
			t.l.Error("查询标签失败", logger.Error(err), logger.Slice[int64]("tag_ids", tagIds))
			return
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		err = t.producer.ProduceSyncEvent(ctx, events.BizTags{
			Biz:   biz,
			BizId: bizId,
			Uid:   uid,
			Tags: transform.SliceFromSlice[domain.Tag, string](tags, func(src domain.Tag) string {
				return src.Name
			}),
		})
		cancel()
		if err != nil {
			t.l.Error("推送事件失败", logger.Error(err), logger.Int64("uid", uid), logger.String("biz", biz), logger.Int64("biz_id", bizId), logger.Slice[int64]("tag_ids", tagIds))
		}
	}()

	return nil
}
