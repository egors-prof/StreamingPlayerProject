package bootstrap

import (
	"context"
	"fmt"

	"github.com/egors-prof/likes_service/internal/adapter/driven/dbstore"
	"github.com/egors-prof/likes_service/internal/config"

	"github.com/egors-prof/likes_service/internal/usecase"
)

func initLayers(cfg config.Config) *App {
	teardown := make([]func(), 0)
	db, err := initDB(*cfg.Postgres)
	if err != nil {
		panic(err)
	}
	storage := dbstore.New(db)
	teardown = append(teardown, func() {
		if err := db.Close(); err != nil {
			fmt.Println(err)
		}
	})
	uc := usecase.New(&cfg, storage)
	httpServer := initHTTPService(&cfg, uc)
	teardown = append(teardown,
		func() {

			ctxShutDown, cancel := context.WithTimeout(context.Background(), gracefulDeadline)
			defer cancel()
			if err := httpServer.Shutdown(ctxShutDown); err != nil {
				return
			}

		},
	)
	return &App{
		cfg:      cfg,
		rest:     httpServer,
		teardown: teardown,
	}
}
