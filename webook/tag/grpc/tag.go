package grpc

import (
	"context"
	tagv1 "webook/api/proto/gen/tag/v1"
	"webook/tag/domain"
	"webook/tag/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
)

type TagServiceServer struct {
	tagv1.UnimplementedTagServiceServer
	svc service.TagService
}

func NewTagGrpcServer(svc service.TagService) *TagServiceServer {
	return &TagServiceServer{svc: svc}
}

func (t *TagServiceServer) Register(server grpc.ServiceRegistrar) {
	tagv1.RegisterTagServiceServer(server, t)
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, req *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	res, err := t.svc.GetBizTags(ctx, req.GetUid(), req.GetBiz(), req.GetBizId())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizTagsResponse{
		Tags: transform.SliceFromSlice[domain.Tag, *tagv1.Tag](res, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) GetTags(ctx context.Context, req *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.svc.GetTags(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagsResponse{
		Tags: transform.SliceFromSlice[domain.Tag, *tagv1.Tag](tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return t.toDTO(src)
		}),
	}, nil
}

func (t *TagServiceServer) AttachTags(ctx context.Context, req *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.svc.AttachTags(ctx, req.GetUid(), req.GetBiz(), req.GetBizId(), req.GetTids())
	return &tagv1.AttachTagsResponse{}, err
}

func (t *TagServiceServer) CreateTag(ctx context.Context, req *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	id, err := t.svc.CreateTag(ctx, req.GetUid(), req.GetName())
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   id,
			Uid:  req.GetUid(),
			Name: req.GetName(),
		},
	}, err
}

func (t *TagServiceServer) toDTO(tag domain.Tag) *tagv1.Tag {
	return &tagv1.Tag{
		Id:   tag.Id,
		Uid:  tag.Uid,
		Name: tag.Name,
	}
}
