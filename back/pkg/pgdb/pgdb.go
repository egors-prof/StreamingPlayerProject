// Package pgdb открывает пул соединений с Postgres.
package pgdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Config читается из переменных окружения с префиксом POSTGRES_.
type Config struct {
	Host            string        `env:"HOST, default=localhost"`
	Port            int           `env:"PORT, default=5432"`
	User            string        `env:"USER, default=postgres"`
	Password        string        `env:"PASSWORD"`
	Database        string        `env:"DATABASE, required"`
	SSLMode         string        `env:"SSL_MODE, default=disable"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNECTIONS, default=25"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNECTIONS, default=25"`
	ConnMaxLifetime time.Duration `env:"CONNECTION_MAX_LIFETIME, default=5m"`
}

// DSN собирает строку подключения в формате URL. Пароль экранируется,
// поэтому спецсимволы в нём не ломают строку.
func (c Config) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:     c.Database,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}
	return u.String()
}

// Connect открывает пул и ждёт, пока база ответит на ping. sqlx.Open сам по
// себе не подключается, поэтому без ping ошибка всплыла бы на первом запросе.
func Connect(ctx context.Context, cfg Config, log *slog.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// При запуске через docker compose база может стартовать позже сервиса.
	const attempts = 10
	for i := 1; ; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			return db, nil
		}
		if i == attempts || ctx.Err() != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ping postgres: %w", err)
		}
		log.Warn("postgres is not ready, retrying", "attempt", i, "error", err)
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			_ = db.Close()
			return nil, ctx.Err()
		}
	}
}

// IsUniqueViolation сообщает, нарушено ли ограничение уникальности (код 23505).
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
