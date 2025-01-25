package ioc

import (
	grpc2 "webook/comment/grpc"

	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/grpcx"
	"github.com/to404hanga/pkg404/logger"
	"google.golang.org/grpc"
)

func InitGrpcxServer(cmt *grpc2.CommentServiceServer, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port int `yaml:"port"`
	}
	s := grpc.NewServer()
	cmt.Register(s)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server: s,
		Port:   cfg.Port,
		Name:   "comment",
		L:      l,
	}
}
