package grpc

import (
	"context"
	"math"
	commentv1 "webook/api/proto/gen/comment/v1"
	"webook/comment/domain"
	"webook/comment/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func NewGrpcServer(svc service.CommentService) *CommentServiceServer {
	return &CommentServiceServer{
		svc: svc,
	}
}

func (c *CommentServiceServer) Register(server grpc.ServiceRegistrar) {
	commentv1.RegisterCommentServiceServer(server, c)
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, req *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	cs, err := c.svc.GetMoreReplies(ctx, req.GetRid(), req.GetMaxId(), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Replies: c.toDTO(cs),
	}, nil
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, req *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	minId := req.GetMinId()
	if minId <= 0 {
		// 第一次查询
		minId = math.MaxInt64
	}
	cs, err := c.svc.GetCommentList(ctx, req.GetBiz(), req.GetBizId(), req.GetMinId(), int(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	return &commentv1.CommentListResponse{
		Comments: c.toDTO(cs),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, req *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	return &commentv1.DeleteCommentResponse{}, c.svc.DeleteComment(ctx, req.GetId())
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, req *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	return &commentv1.CreateCommentResponse{}, c.svc.CreateComment(ctx, c.toDomain(req.GetComment()))
}

func (c *CommentServiceServer) toDTO(cs []domain.Comment) []*commentv1.Comment {
	rpcComments := transform.SliceFromSlice[domain.Comment, *commentv1.Comment](cs, func(cmt domain.Comment) *commentv1.Comment {
		rpcComment := &commentv1.Comment{
			Id:         cmt.Id,
			Uid:        cmt.Commentator.Id,
			Biz:        cmt.Biz,
			BizId:      cmt.BizId,
			Content:    cmt.Content,
			CreateTime: timestamppb.New(cmt.CreateTime),
			UpdateTime: timestamppb.New(cmt.UpdateTime),
		}
		if cmt.RootComment != nil {
			rpcComment.RootComment = &commentv1.Comment{
				Id: cmt.RootComment.Id,
			}
		}
		if cmt.ParentComment != nil {
			rpcComment.ParentComment = &commentv1.Comment{
				Id: cmt.ParentComment.Id,
			}
		}
		return rpcComment
	})
	rpcCommentMap := transform.MapFromSlice[*commentv1.Comment, int64, *commentv1.Comment](rpcComments, func(i int, cmt *commentv1.Comment) (int64, *commentv1.Comment) {
		return cmt.Id, cmt
	})
	// TODO stl map 包编写完毕后改为对应的 ForEach 方法
	for _, cmt := range cs {
		rpcComment := rpcCommentMap[cmt.Id]
		if cmt.RootComment != nil {
			val, ok := rpcCommentMap[cmt.RootComment.Id]
			if ok {
				rpcComment.RootComment = val
			}
		}
		if cmt.ParentComment != nil {
			val, ok := rpcCommentMap[cmt.ParentComment.Id]
			if ok {
				rpcComment.ParentComment = val
			}
		}
	}
	return rpcComments
}

func (c *CommentServiceServer) toDomain(cmt *commentv1.Comment) domain.Comment {
	res := domain.Comment{
		Id:      cmt.GetId(),
		Biz:     cmt.GetBiz(),
		BizId:   cmt.GetBizId(),
		Content: cmt.GetContent(),
		Commentator: domain.User{
			Id: cmt.GetUid(),
		},
	}
	if cmt.GetParentComment() != nil {
		res.ParentComment = &domain.Comment{
			Id: cmt.GetParentComment().GetId(),
		}
	}
	if cmt.GetRootComment() != nil {
		res.RootComment = &domain.Comment{
			Id: cmt.GetRootComment().GetId(),
		}
	}
	return res
}
