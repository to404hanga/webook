package ioc

import (
	"webook/feed/events"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"github.com/to404hanga/pkg404/saramax"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewConsumers(article *events.ArticleEventConsumer, feed *events.FeedEventConsumer) []saramax.Consumer {
	return []saramax.Consumer{
		article,
		feed,
	}
}
