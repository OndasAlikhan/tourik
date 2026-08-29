package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Port             string `env:"PORT"`
	DatabaseName     string `env:"DATABASE_NAME"`
	DatabaseHost     string `env:"DATABASE_HOST"`
	DatabasePort     string `env:"DATABASE_PORT"`
	DatabaseUser     string `env:"DATABASE_USER"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`

	KafkaBroker     string        `env:"KAFKA_BROKER"`
	OutboxTopic     string        `env:"OUTBOX_KAFKA_TOPIC"`
	OutboxPeriod    time.Duration `env:"OUTBOX_WORKER_PERIOD" env-default:"5s"`
	OutboxBatchSize int           `env:"OUTBOX_BATCH_SIZE" env-default:"100"`
}

func NewAppConfig() AppConfig {
	var cfg AppConfig
	cleanenv.ReadEnv(&cfg)

	return cfg
}
