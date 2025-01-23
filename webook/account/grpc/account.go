package grpc

import (
	"context"
	"webook/account/domain"
	"webook/account/service"
	accountv1 "webook/api/proto/gen/account/v1"

	"github.com/to404hanga/pkg404/stl/transform"
	"google.golang.org/grpc"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Register(server *grpc.Server) {
	accountv1.RegisterAccountServiceServer(server, a)
	accountv1.RegisterAccountServiceServer(server, a)
}

func (a *AccountServiceServer) Credit(ctx context.Context, req *accountv1.CreditRequest) (*accountv1.CreditResponse, error) {
	err := a.svc.Credit(ctx, a.toDomain(req))
	return &accountv1.CreditResponse{}, err
}

func (a *AccountServiceServer) toDomain(c *accountv1.CreditRequest) domain.Credit {
	return domain.Credit{
		Biz:   c.GetBiz(),
		BizId: c.GetBizId(),
		Items: transform.SliceFromSlice(c.Items, func(src *accountv1.CreditItem) domain.CreditItem {
			return a.itemToDomain(src)
		}),
	}
}

func (a *AccountServiceServer) itemToDomain(item *accountv1.CreditItem) domain.CreditItem {
	return domain.CreditItem{
		Account:     item.Account,
		Amt:         item.Amt,
		Uid:         item.Uid,
		AccountType: domain.AccountType(item.AccountType),
		Currency:    item.Currency,
	}
}
