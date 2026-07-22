package config

import "github.com/ilyakaznacheev/cleanenv"

type AppConfig struct {
	Port             string `env:"PORT"`
	DatabaseName     string `env:"DATABASE_NAME"`
	DatabaseHost     string `env:"DATABASE_HOST"`
	DatabasePort     string `env:"DATABASE_PORT"`
	DatabaseUser     string `env:"DATABASE_USER"`
	DatabasePassword string `env:"DATABASE_PASSWORD"`
}

func NewAppConfig() AppConfig {
	var cfg AppConfig
	cleanenv.ReadEnv(&cfg)

	return cfg
}
