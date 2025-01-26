package grpc

import (
	"context"
	followv1 "webook/api/proto/gen/follow/v1"
	"webook/follow/domain"
	"webook/follow/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
)

type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowRelationService
}

func NewFollowRelationServiceServer(svc service.FollowRelationService) *FollowServiceServer {
	return &FollowServiceServer{
		svc: svc,
	}
}

func (f *FollowServiceServer) Register(server *grpc.Server) {
	followv1.RegisterFollowServiceServer(server, f)
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, req *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	relationList, err := f.svc.GetFollowee(ctx, req.GetFollower(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}
	return &followv1.GetFolloweeResponse{
		FollowRelations: transform.SliceFromSlice[domain.FollowRelation, *followv1.FollowRelation](relationList, func(fr domain.FollowRelation) *followv1.FollowRelation {
			return f.convertToView(fr)
		}),
	}, nil
}

func (f *FollowServiceServer) FollowInfo(ctx context.Context, req *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	info, err := f.svc.FollowInfo(ctx, req.GetFollower(), req.GetFollowee())
	if err != nil {
		return nil, err
	}
	return &followv1.FollowInfoResponse{
		FollowRelation: f.convertToView(info),
	}, nil
}

func (f *FollowServiceServer) Follow(ctx context.Context, req *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	err := f.svc.Follow(ctx, req.GetFollower(), req.GetFollowee())
	return &followv1.FollowResponse{}, err
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, req *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	err := f.svc.CancelFollow(ctx, req.GetFollower(), req.GetFollowee())
	return &followv1.CancelFollowResponse{}, err
}

func (f *FollowServiceServer) convertToView(relation domain.FollowRelation) *followv1.FollowRelation {
	return &followv1.FollowRelation{
		Follower: relation.Follower,
		Followee: relation.Followee,
	}
}
