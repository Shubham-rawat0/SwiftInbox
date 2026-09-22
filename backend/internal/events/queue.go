package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/utils"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const WebhookEventsQueue = "webhook-events"
const WebhookEventsDeadLetterQueue = "webhook-events-dead-letter"
const WebhookEventsRetryQueue = "webhook-events-retry"

// maxWebhookDeliveryAttempts is the total number of delivery attempts per
// event (1 initial + 4 retried). After the last attempt fails the event is
// parked in the dead-letter queue.
const maxWebhookDeliveryAttempts = 5

// attemptsHeader carries the number of attempts already consumed. It travels
// with the message across retry redeliveries and process restarts, so the
// attempt count survives independently of consumer memory.
const attemptsHeader = "x-attempts"

// webhookRetryDelaysMs maps a failed attempt number to the delay applied
// before the next attempt: after attempt 1 -> 10s, attempt 2 -> 30s, attempt
// 3 -> 2m, attempt 4 -> 10m. Delays are enforced through the per-message TTL
// of the retry queue, so the consumer never blocks or sleeps.
var webhookRetryDelaysMs = []int64{
	10_000,
	30_000,
	120_000,
	600_000,
}

const (
	consumerWorkers  = 8
	consumerPrefetch = consumerWorkers
)

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

// amqpChannel is the subset of *amqp.Channel the queue package needs. Using an
// interface lets the delivery logic be unit-tested without a live broker.
type amqpChannel interface {
	Close() error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Qos(prefetchCount, prefetchSize int, global bool) error //quality of service, prefetch size is max unacknowledge msg a worker can have
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type RabbitMQ struct {
	Conn               *amqp.Connection
	Ch                 amqpChannel
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

// AddRetryQueue declares a queue that holds failed webhook deliveries until
// their per-message TTL expires, then dead-letters them back into targetQueue
// so a consumer can pick them up for another attempt.
func (r *RabbitMQ) AddRetryQueue(name, targetQueue string) error {
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": targetQueue,
	}
	_, err := r.Ch.QueueDeclare(name, true, false, false, false, args)
	if err != nil {
		log.Printf("[WEBHOOK QUEUE] declare failed queue=%s error=%v", name, err)
		return err
	}
	log.Printf("[WEBHOOK QUEUE] declared retry queue=%s dlx=routing://%s durable=true", name, targetQueue)
	return nil
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
	if err := r.Ch.Qos(consumerPrefetch, 0, false); err != nil {
		log.Printf("[WEBHOOK CONSUMER] qos failed: %v", err)
		return err
	}

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

	log.Printf("[WEBHOOK CONSUMER] listening on %s workers=%d", WebhookEventsQueue, consumerWorkers)

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, consumerWorkers)
	for i := 0; i < consumerWorkers; i++ {
		go func() {
			errCh <- r.consumeLoop(workerCtx, msgs)
		}()
	}

	var firstErr error
	for i := 0; i < consumerWorkers; i++ {
		if err := <-errCh; err != nil && firstErr == nil {
			firstErr = err
			cancel()
		}
	}
	if firstErr != nil {
		log.Printf("[WEBHOOK CONSUMER] stopped unexpectedly: %v", firstErr)
	}
	return firstErr
}

func (r *RabbitMQ) consumeLoop(ctx context.Context, msgs <-chan amqp.Delivery) error {
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

	attempt := headerInt(delivery.Headers, attemptsHeader) + 1
	if attempt > maxWebhookDeliveryAttempts {
		return r.deadLetter(ctx, delivery, "delivery retries exhausted", errors.New("attempt limit exceeded"), &event, maxWebhookDeliveryAttempts)
	}

	err = deliverOnce(ctx, client, &event, body, secret)
	if err == nil {
		log.Printf("[WEBHOOK CONSUMER] delivered event=%s webhook=%s attempt=%d", event.Event, event.WebhookID, attempt)
		if err := delivery.Ack(false); err != nil {
			log.Printf("[WEBHOOK CONSUMER] acknowledge failed event=%s webhook=%s delivery_tag=%d error=%v", event.Event, event.WebhookID, delivery.DeliveryTag, err)
			return err
		}
		log.Printf("[WEBHOOK CONSUMER] acknowledged event=%s webhook=%s delivery_tag=%d", event.Event, event.WebhookID, delivery.DeliveryTag)
		return nil
	}

	log.Printf("[WEBHOOK CONSUMER] delivery failed event=%s webhook=%s attempt=%d/%d error=%v", event.Event, event.WebhookID, attempt, maxWebhookDeliveryAttempts, err)

	if attempt >= maxWebhookDeliveryAttempts {
		return r.deadLetter(ctx, delivery, "delivery retries exhausted", err, &event, attempt)
	}

	return r.scheduleRetry(ctx, delivery, &event, attempt)
}

func deliverOnce(ctx context.Context, client *http.Client, event *WebhookEvent, body []byte, secret string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, event.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event.Event)
	req.Header.Set("X-Webhook-Signature", utils.Sign(body, secret))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	return errors.New(resp.Status)
}

// scheduleRetry publishes the failed delivery to the retry queue with a
// per-message TTL equal to the backoff delay and an incremented attempt header.
// The original message is acknowledged only after the publishing succeeds; on
// failure it is left unacknowledged so the broker requeues it for redelivery.
func (r *RabbitMQ) scheduleRetry(ctx context.Context, delivery amqp.Delivery, event *WebhookEvent, attempt int) error {
	delayMs, ok := retryDelayMs(attempt)
	if !ok {
		return r.deadLetter(ctx, delivery, "delivery retries exhausted", fmt.Errorf("no retry delay configured for attempt %d", attempt), event, attempt)
	}

	headers := copyHeaders(delivery.Headers)
	headers[attemptsHeader] = attempt

	publish := amqp.Publishing{
		ContentType:  delivery.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
		Expiration:   strconv.FormatInt(delayMs, 10),
		Body:         delivery.Body,
	}

	log.Printf("[WEBHOOK CONSUMER] scheduling retry event=%s webhook=%s attempt=%d delay=%s", event.Event, event.WebhookID, attempt, time.Duration(delayMs)*time.Millisecond)

	if err := r.Ch.PublishWithContext(ctx, "", WebhookEventsRetryQueue, false, false, publish); err != nil {
		log.Printf("[WEBHOOK CONSUMER] retry publish failed queue=%s delivery_tag=%d error=%v", WebhookEventsRetryQueue, delivery.DeliveryTag, err)
		return err
	}

	if err := delivery.Ack(false); err != nil {
		log.Printf("[WEBHOOK CONSUMER] retry acknowledge failed delivery_tag=%d error=%v", delivery.DeliveryTag, err)
		return err
	}

	log.Printf("[WEBHOOK CONSUMER] scheduled retry queue=%s delivery_tag=%d attempt=%d delay=%s", WebhookEventsRetryQueue, delivery.DeliveryTag, attempt, time.Duration(delayMs)*time.Millisecond)
	return nil
}

// retryDelayMs returns the delay to apply after the given attempt number
// failed, and whether a delay is configured for it.
func retryDelayMs(attempt int) (int64, bool) {
	if attempt < 1 || attempt > len(webhookRetryDelaysMs) {
		return 0, false
	}
	return webhookRetryDelaysMs[attempt-1], true
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
	headers := copyHeaders(delivery.Headers)
	headers["x-dead-letter-reason"] = reason
	headers[attemptsHeader] = attempts

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

func headerInt(headers amqp.Table, key string) int {
	value, ok := headers[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 0
}

func copyHeaders(headers amqp.Table) amqp.Table {
	cloned := make(amqp.Table, len(headers)+1)
	for key, value := range headers {
		cloned[key] = value
	}
	return cloned
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
