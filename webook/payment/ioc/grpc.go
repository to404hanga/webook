package ioc

import (
	grpc2 "webook/payment/grpc"

	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/grpcx"
	ilogger "github.com/to404hanga/pkg404/grpcx/interceptor/logger"
	"github.com/to404hanga/pkg404/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGrpcServer(wechatSvc *grpc2.WechatServiceServer, ecli *clientv3.Client, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdTTL  int64  `yaml:"etcdTTL"`
		EtcdAddr string `yaml:"etcdAddr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(ilogger.NewInterceptorBuilder(l).BuildServerUnaryInterceptor()))
	wechatSvc.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "payment",
		L:          l,
		EtcdTTL:    cfg.EtcdTTL,
		EtcdClient: ecli,
	}
}
