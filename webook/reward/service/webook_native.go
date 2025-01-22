package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	accountv1 "webook/api/proto/gen/account/v1"
	pmtv1 "webook/api/proto/gen/payment/v1"
	"webook/reward/domain"
	"webook/reward/repository"

	"github.com/to404hanga/pkg404/logger"
)

type WechatNativeRewardService struct {
	client pmtv1.WechatPaymentServiceClient
	repo   repository.RewardRepository
	l      logger.Logger
	acli   accountv1.AccountServiceClient
}

var _ RewardService = (*WechatNativeRewardService)(nil)

func NewWechatNativeRewardService(client pmtv1.WechatPaymentServiceClient, repo repository.RewardRepository, l logger.Logger, acli accountv1.AccountServiceClient) RewardService {
	return &WechatNativeRewardService{
		client: client,
		repo:   repo,
		l:      l,
		acli:   acli,
	}
}

func (s *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	res, err := s.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return res, err
	}

	r.Status = domain.RewardStatusInit
	rid, err := s.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	pmtResp, err := s.client.NativePrepay(ctx, &pmtv1.PrepayRequest{
		Amt: &pmtv1.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNo:  s.bizTradeNO(rid),
		Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: pmtResp.GetCodeUrl(),
	}

	go func() {
		err = s.repo.CachedCodeURL(ctx, cu, r)
		if err != nil {
			s.l.Error("缓存二维码失败", logger.Error(err), logger.Int64("rid", rid))
		}
	}()

	return cu, nil
}

func (s *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := s.toRid(bizTradeNO)
	err := s.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}

	if status == domain.RewardStatusPayed {
		r, err := s.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}

		// 平台抽成
		weAmt := int64(float64(r.Amt) * 0.1)
		_, err = s.acli.Credit(ctx, &accountv1.CreditRequest{
			Biz:   "reward",
			BizId: rid,
			Items: []*accountv1.CreditItem{
				{
					AccountType: accountv1.AccountType_AccountTypeSystem,
					Amt:         weAmt,
					Currency:    "CNY",
				},
				{
					Account:     r.Uid,
					Uid:         r.Uid,
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         r.Amt - weAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			s.l.Error("入账失败", logger.Error(err), logger.String("biz_trade_no", bizTradeNO))
			return err
		}
	}
	return nil
}

func (s *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径
	res, err := s.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	// 确保是自己打赏
	if res.Uid != uid {
		return domain.Reward{}, errors.New("非法访问别人的打赏记录")
	}
	// 降级或限流，不走慢路径
	if ctx.Value("limited") == "true" {
		return res, nil
	}
	if !res.Completed() {
		pmtResp, err := s.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
			BizTradeNo: s.bizTradeNO(rid),
		})
		if err != nil {
			s.l.Error("慢路径查询支付状态失败", logger.Error(err), logger.Int64("rid", rid))
			return res, nil
		}
		switch pmtResp.Status {
		case pmtv1.PaymentStatus_PaymentStatusSuccess:
			res.Status = domain.RewardStatusPayed
		case pmtv1.PaymentStatus_PaymentStatusInit:
			res.Status = domain.RewardStatusInit
		case pmtv1.PaymentStatus_PaymentStatusRefund, pmtv1.PaymentStatus_PaymentStatusFailed:
			res.Status = domain.RewardStatusFailed
		default:
			res.Status = domain.RewardStatusUnknown
		}
		err = s.UpdateReward(ctx, s.bizTradeNO(rid), res.Status)
		if err != nil {
			s.l.Error("慢路径更新本地状态失败", logger.Error(err), logger.Int64("rid", rid))
		}
	}
	return res, nil
}

func (s *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (s *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
