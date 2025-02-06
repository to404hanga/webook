package grpc

import (
	"context"
	"encoding/json"
	"time"
	feedv1 "webook/api/proto/gen/feed/v1"
	"webook/feed/domain"
	"webook/feed/service"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
)

type FeedServiceServer struct {
	feedv1.UnimplementedFeedServiceServer
	svc service.FeedService
}

func NewFeedServiceServer(svc service.FeedService) *FeedServiceServer {
	return &FeedServiceServer{
		svc: svc,
	}
}

func (f *FeedServiceServer) Register(server grpc.ServiceRegistrar) {
	feedv1.RegisterFeedServiceServer(server, f)
}

func (f *FeedServiceServer) CreateFeedEvent(ctx context.Context, req *feedv1.CreateFeedEventRequest) (*feedv1.CreateFeedEventResponse, error) {
	err := f.svc.CreateFeedEvent(ctx, f.convertToDomain(req.GetFeedEvent()))
	return &feedv1.CreateFeedEventResponse{}, err
}

func (f *FeedServiceServer) FindFeedEvents(ctx context.Context, req *feedv1.FindFeedEventsRequest) (*feedv1.FindFeedEventsResponse, error) {
	events, err := f.svc.GetFeedEventList(ctx, req.GetUid(), req.GetTimestamp(), int(req.GetLimit()))
	if err != nil {
		return &feedv1.FindFeedEventsResponse{}, err
	}
	res := transform.SliceFromSlice[domain.FeedEvent, *feedv1.FeedEvent](events, func(fe domain.FeedEvent) *feedv1.FeedEvent {
		return f.convertToView(fe)
	})
	return &feedv1.FindFeedEventsResponse{
		FeedEvents: res,
	}, nil
}

func (f *FeedServiceServer) convertToDomain(event *feedv1.FeedEvent) domain.FeedEvent {
	ext := map[string]string{}
	_ = json.Unmarshal([]byte(event.GetContent()), &ext)
	return domain.FeedEvent{
		Id:         event.GetId(),
		CreateTime: time.Unix(event.GetCreateTime(), 0),
		Type:       event.GetType(),
		Ext:        ext,
	}
}

func (f *FeedServiceServer) convertToView(event domain.FeedEvent) *feedv1.FeedEvent {
	val, _ := json.Marshal(event.Ext)
	return &feedv1.FeedEvent{
		Id:         event.Id,
		Type:       event.Type,
		Content:    string(val),
		CreateTime: event.CreateTime.Unix(),
	}
}
