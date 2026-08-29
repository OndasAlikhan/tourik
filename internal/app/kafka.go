package app

import (
	"github.com/segmentio/kafka-go"

	"github.com/OndasAlikhan/tourik/internal/config"
)

func NewKafkaWriter(cfg config.AppConfig) *kafka.Writer {
	return &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBroker),
		Balancer:               &kafka.Hash{},
		AllowAutoTopicCreation: true,
	}
}
