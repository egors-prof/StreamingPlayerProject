package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/pkg/httpx"
	"github.com/egors-prof/StreamingPlayerProject/pkg/pgdb"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/adapter/driven/dbstore"
	httpapi "github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/adapter/driving/http"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/config"
	"github.com/egors-prof/StreamingPlayerProject/services/catalog/internal/usecase/catalog"
)

const shutdownTimeout = 10 * time.Second

func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	db, err := pgdb.Connect(ctx, cfg.Postgres, log)
	if err != nil {
		return err
	}
	defer db.Close()

	uc := catalog.New(dbstore.NewSongRepository(db))
	handler, err := httpapi.New(cfg.TrustedProxies, uc, log)
	if err != nil {
		return err
	}
	if err := httpx.Run(ctx, httpx.NewServer(":"+cfg.HTTPPort, handler), log, shutdownTimeout); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
