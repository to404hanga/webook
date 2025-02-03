package grpc

import (
	"context"
	cronjobv1 "webook/api/proto/gen/cronjob/v1"
	"webook/cronjob/domain"
	"webook/cronjob/service"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CronJobServiceServer struct {
	cronjobv1.UnimplementedCronJobServiceServer
	svc service.CronJobService
}

func NewCronJobServiceServer(svc service.CronJobService) *CronJobServiceServer {
	return &CronJobServiceServer{
		svc: svc,
	}
}

func (s *CronJobServiceServer) Register(server grpc.ServiceRegistrar) {
	cronjobv1.RegisterCronJobServiceServer(server, s)
}

func (s *CronJobServiceServer) AddJob(ctx context.Context, req *cronjobv1.AddJobRequest) (*cronjobv1.AddJobResponse, error) {
	return &cronjobv1.AddJobResponse{}, s.svc.AddJob(ctx, s.convertToDomain(req.GetCronjob()))
}

func (s *CronJobServiceServer) Preempt(ctx context.Context, req *cronjobv1.PreemptRequest) (*cronjobv1.PreemptResponse, error) {
	job, err := s.svc.Preempt(ctx)
	return &cronjobv1.PreemptResponse{
		Cronjob: s.convertToV(job),
	}, err
}

func (s *CronJobServiceServer) ResetNextTime(ctx context.Context, req *cronjobv1.ResetNextTimeRequest) (*cronjobv1.ResetNextTimeResponse, error) {
	err := s.svc.ResetNextTime(ctx, s.convertToDomain(req.GetCronjob()))
	return &cronjobv1.ResetNextTimeResponse{}, err
}

func (s *CronJobServiceServer) convertToV(job domain.CronJob) *cronjobv1.CronJob {
	return &cronjobv1.CronJob{
		Id:         job.Id,
		Name:       job.Name,
		Executor:   job.Executor,
		Cfg:        job.Cfg,
		Expression: job.Expression,
		NextTime:   timestamppb.New(job.NextTime),
	}
}

func (s *CronJobServiceServer) convertToDomain(job *cronjobv1.CronJob) domain.CronJob {
	res := domain.CronJob{}
	if job != nil {
		res.Id = job.GetId()
		res.Name = job.GetName()
		res.Executor = job.GetExecutor()
		res.Cfg = job.GetCfg()
		res.Expression = job.GetExpression()
		res.NextTime = job.GetNextTime().AsTime()
	}
	return res
}
