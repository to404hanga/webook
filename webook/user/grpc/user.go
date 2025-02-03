package grpc

import (
	"context"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/user/domain"
	"webook/user/service"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServiceServer struct {
	userv1.UnimplementedUserServiceServer
	svc service.UserService
}

func NewUserServiceServer(svc service.UserService) *UserServiceServer {
	return &UserServiceServer{svc: svc}
}

func (u *UserServiceServer) Register(server grpc.ServiceRegistrar) {
	userv1.RegisterUserServiceServer(server, u)
}

func (u *UserServiceServer) FindOrCreate(ctx context.Context, req *userv1.FindOrCreateRequest) (*userv1.FindOrCreateResponse, error) {
	user, err := u.svc.FindOrCreate(ctx, req.GetPhone())
	return &userv1.FindOrCreateResponse{
		User: u.convertToV(user),
	}, err
}

func (u *UserServiceServer) FindOrCreateByWechat(ctx context.Context, req *userv1.FindOrCreateByWechatRequest) (*userv1.FindOrCreateByWechatResponse, error) {
	infoV := req.GetInfo()
	user, err := u.svc.FindOrCreateByWechat(ctx, domain.WechatInfo{
		OpenId:  infoV.GetOpenId(),
		UnionId: infoV.GetUnionId(),
	})
	return &userv1.FindOrCreateByWechatResponse{
		User: u.convertToV(user),
	}, err
}

func (u *UserServiceServer) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, err := u.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	return &userv1.LoginResponse{
		User: u.convertToV(user),
	}, err
}

func (u *UserServiceServer) Profile(ctx context.Context, req *userv1.ProfileRequest) (*userv1.ProfileResponse, error) {
	user, err := u.svc.Profile(ctx, req.GetId())
	return &userv1.ProfileResponse{
		User: u.convertToV(user),
	}, err
}

func (u *UserServiceServer) Signup(ctx context.Context, req *userv1.SignupRequest) (*userv1.SignupResponse, error) {
	err := u.svc.SignUp(ctx, u.convertToDomain(req.GetUser()))
	return &userv1.SignupResponse{}, err
}

func (u *UserServiceServer) UpdateNonSensitiveInfo(ctx context.Context, req *userv1.UpdateNonSensitiveInfoRequest) (*userv1.UpdateNonSensitiveInfoResponse, error) {
	err := u.svc.UpdateNonSensitiveInfo(ctx, u.convertToDomain(req.GetUser()))
	return &userv1.UpdateNonSensitiveInfoResponse{}, err
}

func (u *UserServiceServer) convertToV(user domain.User) *userv1.User {
	return &userv1.User{
		Id:         user.Id,
		Email:      user.Email,
		Nickname:   user.Nickname,
		Password:   user.Password,
		Phone:      user.Phone,
		AboutMe:    user.AboutMe,
		CreateTime: timestamppb.New(user.CreateTime),
		Birthday:   timestamppb.New(user.Birthday),
		WechatInfo: &userv1.WechatInfo{
			OpenId:  user.WechatInfo.OpenId,
			UnionId: user.WechatInfo.UnionId,
		},
	}
}

func (u *UserServiceServer) convertToDomain(user *userv1.User) domain.User {
	res := domain.User{}
	if user != nil {
		res.Id = user.GetId()
		res.Email = user.GetEmail()
		res.Nickname = user.GetNickname()
		res.Password = user.GetPassword()
		res.Phone = user.GetPhone()
		res.AboutMe = user.GetAboutMe()
		res.CreateTime = user.GetCreateTime().AsTime()
		res.Birthday = user.GetBirthday().AsTime()
		if user.WechatInfo != nil {
			res.WechatInfo = domain.WechatInfo{
				OpenId:  user.GetWechatInfo().GetOpenId(),
				UnionId: user.GetWechatInfo().GetUnionId(),
			}
		}
	}
	return res
}
