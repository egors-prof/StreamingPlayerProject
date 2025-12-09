package broker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/egors-prof/auth_service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessagePublisher struct {
	channel *amqp.Channel
	queue   *amqp.Queue
}

func New(channel *amqp.Channel, queue *amqp.Queue) *MessagePublisher {
	return &MessagePublisher{
		channel: channel,
		queue:   queue,
	}
}

func (publisher *MessagePublisher) PublishMessage(message domain.Message) error {
	// Отправка сообщения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	body, err := json.Marshal(&Message{
		message.Recipient,
		message.Subject,
		message.Body,
	})
	err = publisher.channel.PublishWithContext(
		ctx,
		"",                     // exchange
		"library-events-queue", // routing key (имя очереди)
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
