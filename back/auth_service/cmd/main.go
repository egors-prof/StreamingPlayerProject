package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	_ "github.com/egors-prof/auth_service/docs"
	"github.com/egors-prof/auth_service/internal/bootstrap"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// @title AuthService API
// @contact.name AuthService API Service
// @contact.url http://test.com
// @contact.email test@test.com
func main() {

	// Load .env file into environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	var cfg config.Config

	err := envconfig.ProcessWith(context.TODO(), &envconfig.Config{Target: &cfg, Lookuper: envconfig.OsLookuper()})
	if err != nil {
		panic(err)
	}

	quitSignal := make(chan os.Signal, 1)
	signal.Notify(quitSignal, os.Interrupt)

	app := bootstrap.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-quitSignal
		cancel()
	}()

	app.Run(ctx)
}
