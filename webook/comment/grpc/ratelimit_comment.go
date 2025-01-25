package grpc

import (
	"context"
	"errors"
	commentv1 "webook/api/proto/gen/comment/v1"
	"webook/comment/service"

	"google.golang.org/grpc"
)

type RateLimitComment struct {
	CommentServiceServer
}

func NewRateLimitGrpcServer(svc service.CommentService) *RateLimitComment {
	return &RateLimitComment{
		CommentServiceServer: CommentServiceServer{
			svc: svc,
		},
	}
}

func (c *RateLimitComment) Register(server grpc.ServiceRegistrar) {
	commentv1.RegisterCommentServiceServer(server, c)
}

func (c *RateLimitComment) GetMoreReplies(ctx context.Context, req *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		return &commentv1.GetMoreRepliesResponse{}, errors.New("资源不足，功能关闭")
	}
	return c.CommentServiceServer.GetMoreReplies(ctx, req)
}

func (c *RateLimitComment) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	// 一般是通过热榜功能，提前计算放到了 Redis 里面，问一下 Redis 就知道是不是热门资源了
	isHotBiz := c.isHotBiz(request.Biz, request.GetBizId())
	if !isHotBiz && ctx.Value("downgrade") == "true" {
		// 非热门资源，触发降级
		return &commentv1.CommentListResponse{}, errors.New("非热门资源被降级")
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitComment) CreateComment(ctx context.Context, req *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrade") == "true" {
		// TODO 转 kafka
		return &commentv1.CreateCommentResponse{}, nil
	}
	return &commentv1.CreateCommentResponse{}, c.svc.CreateComment(ctx, c.toDomain(req.GetComment()))
}

// TODO 补充从 redis 中获取热榜的逻辑
func (c *RateLimitComment) isHotBiz(biz string, bizId int64) bool {
	return true
}
