package bootstrap

import (
	"fmt"
	"net/http"

	http2 "github.com/egors-prof/streaming/internal/adapter/driving/http"
	"github.com/egors-prof/streaming/internal/config"
	"github.com/egors-prof/streaming/internal/usecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func initDB(cfg config.Postgres) (*sqlx.DB, error) {
	postgresOpen := fmt.Sprintf(
		`host=%s
			user=%s
			password=%s
			dbname=%s
			sslmode=disable`,
		cfg.PostgresHost,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDatabase,
	)
	db, err := sqlx.Open("postgres", postgresOpen)
	if err != nil {
		return nil, err
	}

	return db, nil
}
func initHTTPService(
	cfg *config.Config,
	uc *usecase.UseCases,
) *http.Server {
	return http2.New(
		cfg,
		uc,
	)
}
