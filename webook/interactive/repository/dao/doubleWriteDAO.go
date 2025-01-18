package dao

import (
	"context"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/to404hanga/pkg404/logger"
	"gorm.io/gorm"
)

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
	l       logger.Logger
}

var _ InteractiveDAO = (*DoubleWriteDAO)(nil)

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstOnly  = "dst_only"
	PatternDstFirst = "dst_first"
)

func NewDoubleWriteDAO(src, dst *gorm.DB, l logger.Logger) *DoubleWriteDAO {
	return &DoubleWriteDAO{
		src:     NewGormInteractiveDAO(src),
		dst:     NewGormInteractiveDAO(dst),
		pattern: atomicx.NewValueOf[string](PatternSrcOnly),
		l:       l,
	}
}

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			d.l.Error("IncrReadCnt 双写写入 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", bizId), logger.Error(err))
		}
		return nil
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			d.l.Error("IncrReadCnt 双写写入 src 失败", logger.String("biz", biz), logger.Int64("biz_id", bizId), logger.Error(err))
		}
		return nil
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
	case PatternSrcFirst:
		err := d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			return err
		}
		err = d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			d.l.Error("BatchIncrReadCnt 双写写入 dst 失败", logger.Slice[string]("biz_slice", bizs), logger.Slice[int64]("biz_id_slice", bizIds), logger.Error(err))
		}
		return nil
	case PatternDstOnly:
		return d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
	case PatternDstFirst:
		err := d.dst.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			return err
		}
		err = d.src.BatchIncrReadCnt(ctx, bizs, bizIds)
		if err != nil {
			d.l.Error("BatchIncrReadCnt 双写写入 src 失败", logger.Slice[string]("biz_slice", bizs), logger.Slice[int64]("biz_id_slice", bizIds), logger.Error(err))
		}
		return nil
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		err := d.src.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = d.dst.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			d.l.Error("InsertLikeInfo 双写写入 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
		}
		return nil
	case PatternDstOnly:
		return d.dst.InsertLikeInfo(ctx, biz, id, uid)
	case PatternDstFirst:
		err := d.dst.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = d.src.InsertLikeInfo(ctx, biz, id, uid)
		if err != nil {
			d.l.Error("InsertLikeInfo 双写写入 src 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
		}
		return nil
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternSrcFirst:
		err := d.src.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = d.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			d.l.Error("DeleteLikeInfo 双写写入 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
		}
		return nil
	case PatternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, id, uid)
	case PatternDstFirst:
		err := d.dst.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return err
		}
		err = d.src.DeleteLikeInfo(ctx, biz, id, uid)
		if err != nil {
			d.l.Error("DeleteLikeInfo 双写写入 src 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
		}
		return nil
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.InsertCollectionBiz(ctx, cb)
	case PatternSrcFirst:
		err := d.src.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		err = d.dst.InsertCollectionBiz(ctx, cb)
		if err != nil {
			d.l.Error("InsertCollectionBiz 双写写入 dst 失败", logger.String("biz", cb.Biz), logger.Int64("biz_id", cb.BizId), logger.Int64("user_id", cb.Uid), logger.Int64("collection_id", cb.Cid), logger.Error(err))
		}
		return nil
	case PatternDstOnly:
		return d.dst.InsertCollectionBiz(ctx, cb)
	case PatternDstFirst:
		err := d.dst.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		err = d.src.InsertCollectionBiz(ctx, cb)
		if err != nil {
			d.l.Error("InsertCollectionBiz 双写写入 src 失败", logger.String("biz", cb.Biz), logger.Int64("biz_id", cb.BizId), logger.Int64("user_id", cb.Uid), logger.Int64("collection_id", cb.Cid), logger.Error(err))
		}
		return nil
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.GetLikeInfo(ctx, biz, id, uid)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetLikeInfo(ctx, biz, id, uid)
	default:
		return UserLikeBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetLikeInfoV1(ctx context.Context, biz string, id int64, uid int64) (UserLikeBiz, error) {
	// INFO 这是自己写的，不建议
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intrSrc, err := d.src.GetLikeInfo(ctx, biz, id, uid)
		if err != nil {
			return UserLikeBiz{}, err
		}
		go func() {
			intrDst, err := d.dst.GetLikeInfo(ctx, biz, id, uid)
			if err != nil {
				if intrSrc != intrDst {
					d.l.Error("GetLikeInfo 双写读取 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
				}
			}
		}()
		return intrSrc, nil
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetLikeInfo(ctx, biz, id, uid)
	default:
		return UserLikeBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.GetCollectInfo(ctx, biz, id, uid)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetCollectInfo(ctx, biz, id, uid)
	default:
		return UserCollectionBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectInfoV1(ctx context.Context, biz string, id int64, uid int64) (UserCollectionBiz, error) {
	// INFO 这是自己写的，不建议
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intrSrc, err := d.src.GetCollectInfo(ctx, biz, id, uid)
		if err != nil {
			return UserCollectionBiz{}, err
		}
		go func() {
			intrDst, err := d.dst.GetCollectInfo(ctx, biz, id, uid)
			if err != nil {
				if intrSrc != intrDst {
					d.l.Error("GetCollectInfo 双写读取 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Int64("user_id", uid), logger.Error(err))
				}
			}
		}()
		return intrSrc, nil
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetCollectInfo(ctx, biz, id, uid)
	default:
		return UserCollectionBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, id int64) (Interactive, error) {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.Get(ctx, biz, id)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetV1(ctx context.Context, biz string, id int64) (Interactive, error) {
	// 不建议
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intrSrc, err := d.src.Get(ctx, biz, id)
		if err != nil {
			return Interactive{}, err
		}
		go func() {
			intrDst, err := d.dst.Get(ctx, biz, id)
			if err != nil {
				if intrSrc != intrDst {
					d.l.Error("Get 双写读取 dst 失败", logger.String("biz", biz), logger.Int64("biz_id", id), logger.Error(err))
				}
			}
		}()
		return intrSrc, nil
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, id)
	default:
		return Interactive{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	// INFO 这是自己写的
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		return d.src.GetByIds(ctx, biz, ids)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetByIds(ctx, biz, ids)
	default:
		return []Interactive{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetByIdsV1(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	// INFO 这是自己写的，不建议
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcFirst, PatternSrcOnly:
		intrSrc, err := d.src.GetByIds(ctx, biz, ids)
		if err != nil {
			return []Interactive{}, err
		}
		go func() {
			intrDst, err := d.dst.GetByIds(ctx, biz, ids)
			if err != nil {
				if !EqualInteractiveSlice(intrSrc, intrDst) {
					d.l.Error("GetByIds 双写读取 dst 失败", logger.String("biz", biz), logger.Slice[int64]("biz_id_slice", ids), logger.Error(err))
				}
			}
		}()
		return intrSrc, nil
	case PatternDstFirst, PatternDstOnly:
		return d.dst.GetByIds(ctx, biz, ids)
	default:
		return []Interactive{}, ErrUnknownPattern
	}
}
