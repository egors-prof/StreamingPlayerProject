package config

import "github.com/egors-prof/streamingplayer/pkg/pgdb"

type Config struct {
	HTTPPort       string   `env:"HTTP_PORT, default=6789"`
	LogLevel       string   `env:"LOG_LEVEL, default=info"`
	TrustedProxies []string `env:"TRUSTED_PROXIES"`
	JWTSecret      string   `env:"AUTH_JWT_SECRET, required"`

	Postgres pgdb.Config `env:", prefix=POSTGRES_"`
}
