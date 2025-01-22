package grpc

import (
	"context"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/payment/domain"
	"webook/payment/service/wechat"

	"google.golang.org/grpc"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

func (s *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, s)
}

func (s *WechatServiceServer) GetPayment(ctx context.Context, req *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := s.svc.GetPayment(ctx, req.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}

func (s *WechatServiceServer) NativePrepay(ctx context.Context, req *pmtv1.PrepayRequest) (*pmtv1.NativePrepayResponse, error) {
	codeURL, err := s.svc.Prepay(ctx, domain.Payment{
		Amt: domain.Amount{
			Currency: req.GetAmt().GetCurrency(),
			Total:    req.GetAmt().GetTotal(),
		},
		BizTradeNO:  req.GetBizTradeNo(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	return &pmtv1.NativePrepayResponse{
		CodeUrl: codeURL,
	}, nil
}
