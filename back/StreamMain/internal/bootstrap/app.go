package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egors-prof/streaming/internal/adapter/driven/dbstore"
	httpapi "github.com/egors-prof/streaming/internal/adapter/driving/http"
	"github.com/egors-prof/streaming/internal/config"
	"github.com/egors-prof/streaming/internal/limiter"
	"github.com/egors-prof/streaming/internal/usecase/streaming"
	"github.com/egors-prof/streamingplayer/pkg/authjwt"
	"github.com/egors-prof/streamingplayer/pkg/httpx"
	"github.com/egors-prof/streamingplayer/pkg/pgdb"
)

// Стримы длинные, поэтому на остановку даём больше времени. Что не успело
// закончиться, обрывается; плеер на фронте при ошибке запрашивает новый
// билет и продолжает с той же позиции через Range.
const shutdownTimeout = 30 * time.Second

func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	jwtManager, err := authjwt.NewManager(cfg.JWTSecret)
	if err != nil {
		return err
	}

	db, err := pgdb.Connect(ctx, cfg.Postgres, log)
	if err != nil {
		return err
	}
	defer db.Close()

	uc := streaming.New(dbstore.NewTrackStorage(db), jwtManager, cfg.Stream.MusicDir, cfg.Stream.TicketTTL, cfg.Stream.PublicPrefix)
	lim := limiter.New(cfg.Stream.MaxConcurrent, cfg.Stream.QueueTimeout)

	handler, err := httpapi.New(cfg.TrustedProxies, uc, lim, jwtManager, log)
	if err != nil {
		return err
	}
	log.Info("stream limiter configured", "max_concurrent", cfg.Stream.MaxConcurrent, "queue_timeout", cfg.Stream.QueueTimeout)
	if err := httpx.Run(ctx, httpx.NewServer(":"+cfg.HTTPPort, handler), log, shutdownTimeout); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
