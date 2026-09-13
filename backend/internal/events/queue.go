package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const WebhookEventsQueue = "webhook-events"

type WebhookEvent struct {
	WebhookID string `json:"webhook_id"`
	URL       string `json:"url"`
	Event     string `json:"event"`
	MessageID string `json:"message_id"`
}

type RabbitMQ struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	if r.Ch != nil {
		_ = r.Ch.Close()
	}
	if r.Conn != nil {
		_ = r.Conn.Close()
	}
}

func (r *RabbitMQ) AddQueue(name string) error {
	_, err := r.Ch.QueueDeclare(name, true, false, false, false, nil)
	return err
}

func (r *RabbitMQ) Publish(ctx context.Context, event WebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.Ch.PublishWithContext(ctx, "", WebhookEventsQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
