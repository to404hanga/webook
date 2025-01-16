package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/repository/cache"
	cachemocks "webook/internal/repository/cache/mocks"
	"webook/internal/repository/dao"
	daomocks "webook/internal/repository/dao/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCachedUserRepository(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)
		ctx      context.Context
		uid      int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindById(gomock.Any(), uid).
					Return(dao.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						AboutMe:  "自我介绍",
						Phone: sql.NullString{
							String: "1234567890",
							Valid:  true,
						},
						CreateTime: nowMs,
						UpdateTime: 102,
					}, nil)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:         123,
					Email:      "123@qq.com",
					Password:   "123456",
					Birthday:   time.UnixMilli(100),
					AboutMe:    "自我介绍",
					Phone:      "1234567890",
					CreateTime: now,
				}).Return(nil)
				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123,
				Email:      "123@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				AboutMe:    "自我介绍",
				Phone:      "1234567890",
				CreateTime: now,
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{
						Id:         123,
						Email:      "123@qq.com",
						Password:   "123456",
						Birthday:   time.UnixMilli(100),
						AboutMe:    "自我介绍",
						Phone:      "1234567890",
						CreateTime: now,
					}, nil)
				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123,
				Email:      "123@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				AboutMe:    "自我介绍",
				Phone:      "1234567890",
				CreateTime: now,
			},
			wantErr: nil,
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindById(gomock.Any(), uid).
					Return(dao.User{}, dao.ErrRecordNotFound)
				return uc, ud
			},
			uid:      123,
			ctx:      context.Background(),
			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				uid := int64(123)
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), uid).
					Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindById(gomock.Any(), uid).
					Return(dao.User{
						Id: uid,
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						AboutMe:  "自我介绍",
						Phone: sql.NullString{
							String: "1234567890",
							Valid:  true,
						},
						CreateTime: 101,
						UpdateTime: 102,
					}, nil)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:         123,
					Email:      "123@qq.com",
					Password:   "123456",
					Birthday:   time.UnixMilli(100),
					AboutMe:    "自我介绍",
					Phone:      "1234567890",
					CreateTime: time.UnixMilli(101),
				}).Return(errors.New("redis错误"))
				return uc, ud
			},
			uid: 123,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123,
				Email:      "123@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				AboutMe:    "自我介绍",
				Phone:      "1234567890",
				CreateTime: time.UnixMilli(101),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			uc, ud := tc.mock(ctrl)
			svc := NewCachedUserRepository(ud, uc)
			user, err := svc.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
