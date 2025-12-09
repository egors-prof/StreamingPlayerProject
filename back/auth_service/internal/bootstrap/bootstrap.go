package bootstrap

import (
	"context"
	"fmt"

	"github.com/egors-prof/auth_service/internal/adapter/driven/broker"
	"github.com/egors-prof/auth_service/internal/adapter/driven/dbstore"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/auth_service/internal/usecase"
)

func initLayers(cfg config.Config) *App {
	teardown := make([]func(), 0)
	//log := logger.New(cfg.LogLevel, config.ServiceLabel, zap.WithCaller(true))

	conn, ch, err := initRabbitMQ(&cfg.RabbitMQ)
	if err != nil {
		panic(err)
	}

	authQueue, err := initAuthQueue(ch)
	authPublisher := broker.New(ch, authQueue)
	if err != nil {
		panic(err)
	}

	teardown = append(teardown, func() {
		if err := ch.Close(); err != nil {
			fmt.Println(err)
		}
		if err := conn.Close(); err != nil {
			fmt.Println(err)
		}
	})

	db, err := initDB(*cfg.Postgres)
	if err != nil {
		panic(err)
	}

	storage := dbstore.New(db)
	//log.Info("Database connection established")

	teardown = append(teardown, func() {
		if err := db.Close(); err != nil {
			fmt.Println(err)
			//log.Error(err.Error())
		}
	})

	uc := usecase.New(cfg, storage, authPublisher)

	httpSrv := initHTTPService(&cfg, uc)

	teardown = append(teardown,
		func() {
			//log.Info("HTTP is shutting down")
			ctxShutDown, cancel := context.WithTimeout(context.Background(), gracefulDeadline)
			defer cancel()
			if err := httpSrv.Shutdown(ctxShutDown); err != nil {
				//log.Error(fmt.Sprintf("server Shutdown Failed:%s", err))
				return
			}
			//log.Info("HTTP is shut down")
		},
	)

	return &App{
		cfg:      cfg,
		rest:     httpSrv,
		teardown: teardown,
	}
}
