package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

const WebhookEventsQueue = "webhook-events"
const WebhookEventsDeadLetterQueue = "webhook-events-dead-letter"

const deliveryRetries = 5
const maxDeliveryAttempts = 1 + deliveryRetries

type Publisher interface {
	Publish(context.Context, WebhookEvent) error
}

type WebhookEvent struct {
	WebhookID       string `json:"webhook_id"`
	MailboxID       string `json:"mailbox_id"`
	URL             string `json:"url"`
	Event           string `json:"event"`
	MessageID       string `json:"message_id"`
	SecretEncrypted string `json:"secret_encrypted"`
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

func (r *RabbitMQ) Consume(ctx context.Context) error {
	msgs, err := r.Ch.Consume(
		WebhookEventsQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("[WEBHOOK CONSUMER] listening on %s", WebhookEventsQueue)
	client := &http.Client{Timeout: 15 * time.Second}
	for {
		select {
		case <-ctx.Done():
			log.Printf("[WEBHOOK CONSUMER] stopped: %v", ctx.Err())
			return nil
		case delivery, ok := <-msgs:
			if !ok {
				return errors.New("webhook consumer channel closed")
			}
			if err := r.handleDelivery(ctx, client, delivery); err != nil {
				log.Printf("[WEBHOOK CONSUMER] delivery handling failed: %v", err)
			}
		}
	}
}

func (r *RabbitMQ) handleDelivery(ctx context.Context, client *http.Client, delivery amqp.Delivery) error {
	var event WebhookEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return r.deadLetter(ctx, delivery, "invalid event payload", err)
	}

	key, err := utils.WebhookEncryptionKey()
	if err != nil {
		return r.deadLetter(ctx, delivery, "invalid webhook encryption key", err)
	}
	secret, err := utils.Decrypt(event.SecretEncrypted, key)
	if err != nil {
		return r.deadLetter(ctx, delivery, "unable to decrypt webhook secret", err)
	}

	body, err := json.Marshal(struct {
		WebhookID string `json:"webhook_id"`
		MailboxID string `json:"mailbox_id"`
		Event     string `json:"event"`
		MessageID string `json:"message_id"`
	}{
		WebhookID: event.WebhookID,
		MailboxID: event.MailboxID,
		Event:     event.Event,
		MessageID: event.MessageID,
	})
	if err != nil {
		return r.deadLetter(ctx, delivery, "unable to encode webhook payload", err)
	}

	for attempt := 1; attempt <= maxDeliveryAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, event.URL, bytes.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Webhook-Event", event.Event)
			req.Header.Set("X-Webhook-Signature", utils.Sign(body, secret))
			resp, requestErr := client.Do(req)
			if requestErr == nil {
				_ = resp.Body.Close()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
					log.Printf("[WEBHOOK CONSUMER] delivered event=%s webhook=%s attempt=%d", event.Event, event.WebhookID, attempt)
					return delivery.Ack(false)
				}
				err = errors.New(resp.Status) 
			} else {
				err = requestErr
			}
		}
		log.Printf("[WEBHOOK CONSUMER] delivery failed event=%s webhook=%s attempt=%d/%d error=%v", event.Event, event.WebhookID, attempt, maxDeliveryAttempts, err)
	}

	return r.deadLetter(ctx, delivery, "delivery retries exhausted", err)
}

func (r *RabbitMQ) deadLetter(ctx context.Context, delivery amqp.Delivery, reason string, cause error) error {
	log.Printf("[WEBHOOK CONSUMER] sending event to dead-letter queue reason=%s error=%v", reason, cause)
	if err := r.Ch.PublishWithContext(ctx, "", WebhookEventsDeadLetterQueue, false, false, amqp.Publishing{
		ContentType:  delivery.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers: amqp.Table{
			"x-dead-letter-reason": reason,
		},
		Body: delivery.Body,
	}); err != nil {
		return err
	}
	return delivery.Ack(false)
}
