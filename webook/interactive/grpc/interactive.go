package grpc

import (
	"context"
	intrv1 "webook/api/proto/gen/intr/v1"
	"webook/interactive/domain"
	"webook/interactive/service"

	"google.golang.org/grpc"
)

type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{
		svc: svc,
	}
}

func (i *InteractiveServiceServer) Register(s *grpc.Server) {
	intrv1.RegisterInteractiveServiceServer(s, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (response *intrv1.IncrReadCntResponse, err error) {
	err = i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}
func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (response *intrv1.LikeResponse, err error) {
	err = i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.LikeResponse{}, err
}
func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (response *intrv1.CancelLikeResponse, err error) {
	err = i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &intrv1.CancelLikeResponse{}, err
}
func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (response *intrv1.CollectResponse, err error) {
	err = i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.CollectResponse{}, err
}
func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (response *intrv1.GetResponse, err error) {
	var intr domain.Interactive
	intr, err = i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return &intrv1.GetResponse{}, err
	}
	return &intrv1.GetResponse{
		Intr: i.toDTO(intr),
	}, nil
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (response *intrv1.GetByIdsResponse, err error) {
	var intrs map[int64]domain.Interactive
	intrs, err = i.svc.GetByIds(ctx, request.GetBiz(), request.GetIds())
	if err != nil {
		return &intrv1.GetByIdsResponse{}, err
	}
	ret := make(map[int64]*intrv1.Interactive, len(intrs))
	for id, intr := range intrs {
		ret[id] = i.toDTO(intr)
	}
	return &intrv1.GetByIdsResponse{
		Intrs: ret,
	}, nil
}

func (i *InteractiveServiceServer) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}
