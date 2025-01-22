package grpc

import (
	"context"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/reward/domain"
	"webook/reward/service"

	"google.golang.org/grpc"
)

type RewardServiceServer struct {
	rewardv1.UnimplementedRewardServiceServer
	svc service.RewardService
}

var _ rewardv1.RewardServiceServer = (*RewardServiceServer)(nil)

func NewRewardServiceServer(svc service.RewardService) *RewardServiceServer {
	return &RewardServiceServer{svc: svc}
}

func (r *RewardServiceServer) Register(server *grpc.Server) {
	rewardv1.RegisterRewardServiceServer(server, r)
}

func (r *RewardServiceServer) PreReward(ctx context.Context, req *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	codeURL, err := r.svc.PreReward(ctx, domain.Reward{
		Uid: req.GetUid(),
		Target: domain.Target{
			Biz:     req.GetBiz(),
			BizId:   req.GetBizId(),
			BizName: req.GetBizName(),
			Uid:     req.GetTargetUid(),
		},
		Amt: req.GetAmt(),
	})
	return &rewardv1.PreRewardResponse{
		CodeUrl: codeURL.URL,
		Rid:     codeURL.Rid,
	}, err
}

func (r *RewardServiceServer) GetReward(ctx context.Context, req *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	rw, err := r.svc.GetReward(ctx, req.GetRid(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &rewardv1.GetRewardResponse{
		Status: rewardv1.RewardStatus(rw.Status),
	}, nil
}
