package ioc

import (
	grpc2 "webook/search/grpc"

	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/grpcx"
	"github.com/to404hanga/pkg404/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGrpcxServer(syncRpc *grpc2.SyncServiceServer, searchRpc *grpc2.SearchServiceServer, ecli *clientv3.Client, l logger.Logger) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
		EtcdTTL  int64  `yaml:"etcdTTL"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	server := grpc.NewServer()
	syncRpc.Register(server)
	searchRpc.Register(server)
	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		EtcdClient: ecli,
		Name:       "search",
		EtcdTTL:    cfg.EtcdTTL,
		L:          l,
	}
}
