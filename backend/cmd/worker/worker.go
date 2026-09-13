package worker

import (
	queue "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/events"
)

func StartWorker(recorder queue.DeadLetterRecorder) (*queue.RabbitMQ, error) {
	rabbit, err := queue.NewRabbitMQ(
		"amqp://guest:guest@localhost:5672/",
	)
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

	return rabbit, nil
}
