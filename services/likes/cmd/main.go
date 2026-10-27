package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/egors-prof/StreamingPlayerProject/pkg/logger"
	_ "github.com/egors-prof/StreamingPlayerProject/services/likes/docs"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/bootstrap"
	"github.com/egors-prof/StreamingPlayerProject/services/likes/internal/config"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// @title       Likes Service API
// @version     2.0
// @description Лайки пользователя. Все ручки требуют access-токен.
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
// @description Формат: "Bearer {access_token}"
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	var cfg config.Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	log := logger.New("likes", cfg.LogLevel)
	if err := bootstrap.Run(ctx, cfg, log); err != nil {
		log.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("service stopped")
}
