package ioc

import (
	smsv1 "webook/api/proto/gen/sms/v1"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitSmsRpcClient() smsv1.SmsServiceClient {
	type Config struct {
		Target string `yaml:"target"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.sms", &cfg)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.NewClient(cfg.Target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := smsv1.NewSmsServiceClient(conn)
	return client
}
