package config

import (
	"fmt"
	"time"
)

const ServiceLabel = "auth_service"

type Config struct {
	HTTPPort   string     `env:"HTTP_PORT" default:"8284"`
	Postgres   *Postgres  `env:",prefix=POSTGRES_"`
	AuthParams AuthParams `env:",prefix=JWT_"`
	RabbitMQ   RabbitMQ   `env:",prefix=RABBITMQ_"`
}

type AuthParams struct {
	AccessTokenTllMinutes int    `env:"ACCESS_TOKEN_TLL_MINUTES"`
	RefreshTokenTllDays   int    `env:"REFRESH_TOKEN_TLL_DAYS"`
	SECRET                string `env:"SECRET"`
}

type Postgres struct {
	PostgresHost          string        `env:"HOST" default:"localhost"`
	PostgresPort          int           `env:"PORT" default:"5433"`
	PostgresUser          string        `env:"USER" default:"postgres"`
	PostgresPassword      string        `env:"PASSWORD" default:"postgres"`
	PostgresDatabase      string        `env:"DATABASE"`
	PostgresSSLMode       string        `env:"SSL_MODE" default:"disable"`
	MaxIdleConnections    int           `env:"MAX_IDLE_CONNECTIONS" default:"25"`
	MaxOpenConnections    int           `env:"MAX_OPEN_CONNECTIONS" default:"25"`
	ConnectionMaxLifetime time.Duration `env:"CONNECTION_MAX_LIFETIME" default:"5m"`
}

type RabbitMQ struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     int    `env:"PORT" default:"5673"`
	User     string `env:"USER" default:"guest"`
	Password string `env:"PASSWORD" default:"guest"`
}

func (c *Postgres) ConnectionURL() string {
	if c.PostgresUser == "" {
		return fmt.Sprintf("host=%s port=%d  dbname=%s sslmode=disable",
			c.PostgresHost, c.PostgresPort, c.PostgresDatabase)
	}

	if c.PostgresPassword == "" {
		return fmt.Sprintf("host=%s port=%d user=%s  dbname=%s sslmode=disable",
			c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresDatabase)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDatabase)
}
