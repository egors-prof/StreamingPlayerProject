// Package bootstrap собирает сервис из слоёв и запускает его.
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egors-prof/auth_service/internal/adapter/driven/broker"
	"github.com/egors-prof/auth_service/internal/adapter/driven/dbstore"
	httpapi "github.com/egors-prof/auth_service/internal/adapter/driving/http"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/auth_service/internal/port/driven"
	"github.com/egors-prof/auth_service/internal/usecase/registration"
	"github.com/egors-prof/auth_service/internal/usecase/session"
	"github.com/egors-prof/streamingplayer/pkg/authjwt"
	"github.com/egors-prof/streamingplayer/pkg/httpx"
	"github.com/egors-prof/streamingplayer/pkg/pgdb"
)

const shutdownTimeout = 10 * time.Second

// Run работает, пока не отменён ctx. Ресурсы закрываются через defer, то есть
// в обратном порядке: сначала останавливается HTTP-сервер (дожидаясь текущих
// запросов), затем брокер, и только потом база.
func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	jwtManager, err := authjwt.NewManager(cfg.Auth.JWTSecret)
	if err != nil {
		return err
	}

	db, err := pgdb.Connect(ctx, cfg.Postgres, log)
	if err != nil {
		return err
	}
	defer db.Close()

	var publisher driven.EventPublisher = broker.Noop{Log: log}
	if cfg.RabbitMQ.Enabled {
		p, err := broker.Connect(ctx, cfg.RabbitMQ.URL(), cfg.RabbitMQ.Queue, log)
		if err != nil {
			return err
		}
		defer p.Close()
		publisher = p
	}

	users := dbstore.NewUserStorage(db)
	tokens := dbstore.NewTokenStorage(db)

	reg := registration.New(users, publisher, log)
	sess := session.New(users, tokens, jwtManager, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)

	handler, err := httpapi.New(cfg, reg, sess, log)
	if err != nil {
		return err
	}
	srv := httpx.NewServer(":"+cfg.HTTPPort, handler)
	if err := httpx.Run(ctx, srv, log, shutdownTimeout); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
