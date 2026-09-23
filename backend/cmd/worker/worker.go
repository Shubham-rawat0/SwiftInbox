package worker

import (
	"os"

	queue "github.com/Shubham-rawat0/temp-mail/SwiftInbox/backend/internal/events"
)

func StartWorker(recorder queue.DeadLetterRecorder) (*queue.RabbitMQ, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}
	rabbit, err := queue.NewRabbitMQ(url)
	if err != nil {
		return nil, err
	}
	rabbit.DeadLetterRecorder = recorder

	if err := rabbit.AddQueue(queue.WebhookEventsQueue); err != nil {
		rabbit.Close()
		return nil, err
	}
	if err := rabbit.AddQueue(queue.WebhookEventsDeadLetterQueue); err != nil {
		rabbit.Close()
		return nil, err
	}
	if err := rabbit.AddRetryQueue(queue.WebhookEventsRetryQueue, queue.WebhookEventsQueue); err != nil {
		rabbit.Close()
		return nil, err
	}

	return rabbit, nil
}
