package ioc

import (
	followv1 "webook/api/proto/gen/follow/v1"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitFollowClient() followv1.FollowServiceClient {
	type Config struct {
		Target string `yaml:"target"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.follow", &cfg)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.NewClient(cfg.Target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := followv1.NewFollowServiceClient(conn)
	return client
}
