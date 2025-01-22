package ioc

import (
	accountv1 "webook/api/proto/gen/account/v1"

	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAccountClient(etcdClient *clientv3.Client) accountv1.AccountServiceClient {
	type Config struct {
		Target string `yaml:"target" json:"target"`
		Secure bool   `yaml:"secure" json:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.account", &cfg)
	if err != nil {
		panic(err)
	}
	rs, err := resolver.NewBuilder(etcdClient)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{
		grpc.WithResolvers(rs),
	}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		panic(err)
	}
	return accountv1.NewAccountServiceClient(cc)
}
