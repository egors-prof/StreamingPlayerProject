package bootstrap

import (
	"fmt"
	"net/http"

	http2 "github.com/egors-prof/auth_service/internal/adapter/driving/http"
	"github.com/egors-prof/auth_service/internal/config"
	"github.com/egors-prof/auth_service/internal/usecase"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func initDB(cfg config.Postgres) (*sqlx.DB, error) {
	connectStr := fmt.Sprintf(
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
	db, err := sqlx.Open("postgres", connectStr)

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

func initRabbitMQ(cfg *config.RabbitMQ) (*amqp.Connection, *amqp.Channel, error) {
	connURL := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	// Connect
	conn, err := amqp.Dial(connURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return conn, ch, nil
}

func initAuthQueue(ch *amqp.Channel) (*amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		"service-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare auth queue: %w", err)
	}

	return &queue, nil
}
