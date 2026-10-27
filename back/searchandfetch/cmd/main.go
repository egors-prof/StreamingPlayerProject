package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/egors-prof/searchService/internal/bootstrap"
	"github.com/egors-prof/searchService/internal/config"
	"github.com/egors-prof/streamingplayer/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	var cfg config.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	log := logger.New("catalog_service", cfg.LogLevel)
	if err := bootstrap.Run(ctx, cfg, log); err != nil {
		log.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("service stopped")
}
