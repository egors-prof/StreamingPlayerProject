package config

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	HTTPPort string    `env:"HTTP_PORT" default:"8888"`
	Postgres *Postgres `env:",prefix=POSTGRES_"`
}

type Postgres struct {
	PostgresHost     string `env:"HOST" `
	PostgresPort     int    `env:"PORT" `
	PostgresUser     string `env:"USER" `
	PostgresPassword string `env:"PASSWORD" `
	PostgresDatabase string `env:"DATABASE" `
	PostgresSSLMode  string `env:"SSL_MODE" `
}

func GetConfigs() (*Config, error) {
	if err := godotenv.Load(); err != nil {

		return &Config{}, err
	}

	var cfg Config

	err := envconfig.ProcessWith(context.TODO(), &envconfig.Config{Target: &cfg, Lookuper: envconfig.OsLookuper()})
	if err != nil {

		return &Config{}, err
	}

	return &cfg, nil
}
