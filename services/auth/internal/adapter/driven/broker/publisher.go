// Package broker публикует доменные события в RabbitMQ.
package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/egors-prof/StreamingPlayerProject/services/auth/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher держит одно соединение и один канал. Канал amqp нельзя
// использовать из нескольких горутин одновременно, поэтому публикация
// защищена мьютексом.
type Publisher struct {
	mu    sync.Mutex
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

// Connect подключается к брокеру с повторами (при старте через docker compose
// RabbitMQ может подняться позже сервиса) и объявляет очередь.
func Connect(ctx context.Context, url, queue string, log *slog.Logger) (*Publisher, error) {
	var conn *amqp.Connection
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Warn("rabbitmq is not ready, retrying", "attempt", attempt, "error", err)
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	// Режим подтверждений: брокер отвечает ack, когда сообщение принято.
	// Без него Publish «успешен», даже если сообщение потерялось по дороге.
	if err := ch.Confirm(false); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("enable confirms: %w", err)
	}
	// Имя очереди одно и для объявления, и для публикации. В старой версии
	// объявлялась "service-queue", а публикация шла в "library-events-queue",
	// и сообщения молча пропадали.
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}
	return &Publisher{conn: conn, ch: ch, queue: queue}, nil
}

func (p *Publisher) PublishUserRegistered(ctx context.Context, e domain.UserRegistered) error {
	body, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	p.mu.Lock()
	defer p.mu.Unlock()

	confirm, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, "", p.queue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent, // сообщение переживёт перезапуск брокера
		ContentType:  "application/json",
		Type:         "user.registered",
		Timestamp:    e.OccurredAt,
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("wait confirm: %w", err)
	}
	if !acked {
		return errors.New("broker rejected the message")
	}
	return nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return errors.Join(p.ch.Close(), p.conn.Close())
}

// Noop используется, когда RabbitMQ выключен (RABBITMQ_ENABLED=false),
// например для запуска сервиса без брокера при локальной разработке.
type Noop struct{ Log *slog.Logger }

func (n Noop) PublishUserRegistered(_ context.Context, e domain.UserRegistered) error {
	n.Log.Info("rabbitmq disabled, event not published", "event", "user.registered", "user_id", e.UserID)
	return nil
}
