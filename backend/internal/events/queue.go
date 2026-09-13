package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const WebhookEventsQueue = "webhook-events"
const WebhookEventsDeadLetterQueue = "webhook-events-dead-letter"

const deliveryRetries = 5
const maxDeliveryAttempts = 1 + deliveryRetries

type Publisher interface {
	Publish(context.Context, WebhookEvent) error
}

type DeadLetterRecorder interface {
	CreateWebhookDeadLetter(context.Context, postgres.CreateWebhookDeadLetterParams) error
}

type WebhookEvent struct {
	WebhookID       string `json:"webhook_id"`
	DeveloperID     string `json:"developer_id"`
	MailboxID       string `json:"mailbox_id"`
	URL             string `json:"url"`
	Event           string `json:"event"`
	MessageID       string `json:"message_id"`
	SecretEncrypted string `json:"secret_encrypted"`
}

type RabbitMQ struct {
	Conn               *amqp.Connection
	Ch                 *amqp.Channel
	DeadLetterRecorder DeadLetterRecorder
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	log.Printf("[WEBHOOK QUEUE] connecting to RabbitMQ")
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] connection failed: %v", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] channel creation failed: %v", err)
		conn.Close()
		return nil, err
	}

	log.Printf("[WEBHOOK QUEUE] connected")

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
	}, nil
}

func (r *RabbitMQ) Close() {
	log.Printf("[WEBHOOK QUEUE] closing RabbitMQ connection")
	if r.Ch != nil {
		_ = r.Ch.Close()
	}
	if r.Conn != nil {
		_ = r.Conn.Close()
	}
}

func (r *RabbitMQ) AddQueue(name string) error {
	_, err := r.Ch.QueueDeclare(name, true, false, false, false, nil)
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] declare failed queue=%s error=%v", name, err)
		return err
	}
	log.Printf("[WEBHOOK QUEUE] declared queue=%s durable=true", name)
	return err
}

func (r *RabbitMQ) Publish(ctx context.Context, event WebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] encode failed event=%s webhook=%s message=%s error=%v", event.Event, event.WebhookID, event.MessageID, err)
		return err
	}

	err = r.Ch.PublishWithContext(ctx, "", WebhookEventsQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] publish failed queue=%s event=%s webhook=%s mailbox=%s message=%s error=%v", WebhookEventsQueue, event.Event, event.WebhookID, event.MailboxID, event.MessageID, err)
		return err
	}

	log.Printf("[WEBHOOK QUEUE] published queue=%s event=%s webhook=%s mailbox=%s message=%s bytes=%d", WebhookEventsQueue, event.Event, event.WebhookID, event.MailboxID, event.MessageID, len(body))
	return nil
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
				log.Printf("[WEBHOOK CONSUMER] delivery channel closed")
				return errors.New("webhook consumer channel closed")
			}
			log.Printf("[WEBHOOK CONSUMER] received queue=%s delivery_tag=%d redelivered=%t bytes=%d", WebhookEventsQueue, delivery.DeliveryTag, delivery.Redelivered, len(delivery.Body))
			if err := r.handleDelivery(ctx, client, delivery); err != nil {
				log.Printf("[WEBHOOK CONSUMER] delivery handling failed: %v", err)
			}
		}
	}
}

func (r *RabbitMQ) handleDelivery(ctx context.Context, client *http.Client, delivery amqp.Delivery) error {
	var event WebhookEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return r.deadLetter(ctx, delivery, "invalid event payload", err, nil, 0)
	}

	key, err := utils.WebhookEncryptionKey()
	if err != nil {
		return r.deadLetter(ctx, delivery, "invalid webhook encryption key", err, &event, 0)
	}
	secret, err := utils.Decrypt(event.SecretEncrypted, key)
	if err != nil {
		return r.deadLetter(ctx, delivery, "unable to decrypt webhook secret", err, &event, 0)
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
		return r.deadLetter(ctx, delivery, "unable to encode webhook payload", err, &event, 0)
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
					if err := delivery.Ack(false); err != nil {
						log.Printf("[WEBHOOK CONSUMER] acknowledge failed event=%s webhook=%s delivery_tag=%d error=%v", event.Event, event.WebhookID, delivery.DeliveryTag, err)
						return err
					}
					log.Printf("[WEBHOOK CONSUMER] acknowledged event=%s webhook=%s delivery_tag=%d", event.Event, event.WebhookID, delivery.DeliveryTag)
					return nil
				}
				err = errors.New(resp.Status)
			} else {
				err = requestErr
			}
		}
		log.Printf("[WEBHOOK CONSUMER] delivery failed event=%s webhook=%s attempt=%d/%d error=%v", event.Event, event.WebhookID, attempt, maxDeliveryAttempts, err)
	}

	return r.deadLetter(ctx, delivery, "delivery retries exhausted", err, &event, maxDeliveryAttempts)
}

func (r *RabbitMQ) deadLetter(
	ctx context.Context,
	delivery amqp.Delivery,
	reason string,
	cause error,
	event *WebhookEvent,
	attempts int,
) error {
	log.Printf(
		"[WEBHOOK CONSUMER] sending event to dead-letter queue reason=%s error=%v",
		reason,
		cause,
	)

	// 1. Record the dead-letter event in Postgres.
	if r.DeadLetterRecorder != nil && event != nil {
		record := postgres.CreateWebhookDeadLetterParams{
			ID:          uuid.New(),
			DeveloperID: parseUUID(event.DeveloperID),
			WebhookID:   parseUUID(event.WebhookID),
			MailboxID:   nullableUUID(event.MailboxID),
			MessageID:   nullableUUID(event.MessageID),
			Event:       event.Event,
			Url:         event.URL,
			Reason:      reason,
			Attempts:    int32(attempts),
		}

		if err := r.DeadLetterRecorder.CreateWebhookDeadLetter(ctx, record); err != nil {
			log.Printf(
				"[WEBHOOK CONSUMER] dead-letter record failed webhook=%s error=%v",
				event.WebhookID,
				err,
			)
			return err
		}
	}

	// 2. Copy the original headers so we don't lose them.
	headers := make(amqp.Table, len(delivery.Headers)+2)

	for key, value := range delivery.Headers {
		headers[key] = value
	}

	headers["x-dead-letter-reason"] = reason
	headers["x-attempts"] = attempts

	// 3. Publish the original message to the dead-letter queue.
	if err := r.Ch.PublishWithContext(
		ctx,
		"",
		WebhookEventsDeadLetterQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  delivery.ContentType,
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Body:         delivery.Body,
		},
	); err != nil {
		log.Printf(
			"[WEBHOOK CONSUMER] dead-letter publish failed queue=%s delivery_tag=%d error=%v",
			WebhookEventsDeadLetterQueue,
			delivery.DeliveryTag,
			err,
		)
		return err
	}

	// 4. Only ACK the original message after DLQ publish succeeds.
	if err := delivery.Ack(false); err != nil {
		log.Printf(
			"[WEBHOOK CONSUMER] dead-letter acknowledge failed delivery_tag=%d error=%v",
			delivery.DeliveryTag,
			err,
		)
		return err
	}

	log.Printf(
		"[WEBHOOK CONSUMER] dead-lettered queue=%s delivery_tag=%d reason=%s attempts=%d",
		WebhookEventsDeadLetterQueue,
		delivery.DeliveryTag,
		reason,
		attempts,
	)

	return nil
}

func parseUUID(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return id
}

func nullableUUID(value string) uuid.NullUUID {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}
