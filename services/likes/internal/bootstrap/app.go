package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/authjwt"
	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/pkg/pgdb"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/adapter/driven/dbstore"
	httpapi "github.com/egors-prof/StreamingPlayerProject/services/likes/internal/adapter/driving/http"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/config"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/usecase/likes"
)

const shutdownTimeout = 10 * time.Second

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

	uc := likes.New(dbstore.NewLikeStorage(db))
	handler, err := httpapi.New(cfg.TrustedProxies, uc, jwtManager, log)
	if err != nil {
		return err
	}
	if err := httpx.Run(ctx, httpx.NewServer(":"+cfg.HTTPPort, handler), log, shutdownTimeout); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
