package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/egors-prof/auth_service/docs"
	"github.com/egors-prof/auth_service/internal/bootstrap"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/streamingplayer/pkg/logger"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// @title       Auth Service API
// @version     2.0
// @description Регистрация, вход, обновление сессии и выход.
// @BasePath    /
func main() {
	// SIGTERM присылают docker stop и Kubernetes, SIGINT — Ctrl+C.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load() // .env удобен локально; в контейнере переменные задаёт compose

	var cfg config.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	log := logger.New("auth_service", cfg.LogLevel)
	if err := bootstrap.Run(ctx, cfg, log); err != nil {
		log.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("service stopped")
}
