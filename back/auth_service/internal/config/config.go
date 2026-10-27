package config

import (
	"fmt"
	"time"

	"github.com/egors-prof/streamingplayer/pkg/pgdb"
)

// Config читается из переменных окружения (go-envconfig).
// Синтаксис тегов: `env:"NAME, default=value"`; prefix задаёт общий префикс
// для вложенной структуры.
type Config struct {
	HTTPPort       string   `env:"HTTP_PORT, default=8284"`
	LogLevel       string   `env:"LOG_LEVEL, default=info"`
	TrustedProxies []string `env:"TRUSTED_PROXIES"`

	Postgres pgdb.Config `env:", prefix=POSTGRES_"`
	Auth     Auth        `env:", prefix=AUTH_"`
	RabbitMQ RabbitMQ    `env:", prefix=RABBITMQ_"`
}

type Auth struct {
	JWTSecret       string        `env:"JWT_SECRET, required"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL, default=15m"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL, default=720h"`
	// Refresh-токен хранится в HttpOnly-cookie: JavaScript его не видит,
	// поэтому при XSS его не украсть. Secure=true требует HTTPS.
	CookieSecure bool   `env:"COOKIE_SECURE, default=true"`
	CookiePath   string `env:"COOKIE_PATH, default=/api/auth"`
}

type RabbitMQ struct {
	Enabled  bool   `env:"ENABLED, default=true"`
	Host     string `env:"HOST, default=localhost"`
	Port     int    `env:"PORT, default=5672"`
	User     string `env:"USER, default=guest"`
	Password string `env:"PASSWORD, default=guest"`
	Queue    string `env:"QUEUE, default=user.registered"`
}

func (r RabbitMQ) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", r.User, r.Password, r.Host, r.Port)
}
