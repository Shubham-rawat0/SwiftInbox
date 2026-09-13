package worker

import (
	"log"

	queue "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/events"
)

func StartWorker(){
	rabbit, err := queue.NewRabbitMQ(
    	"amqp://guest:guest@localhost:5672/",
	)

	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = rabbit.Ch.QueueDeclare(
    "webhook-events",
    true,  // durable
    false, // delete when unused
    false, // exclusive
    false, // no-wait
    nil,)

	if err != nil {
		log.Fatal(err)
	}

	defer rabbit.Close()
}