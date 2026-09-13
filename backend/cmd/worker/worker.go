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

defer rabbit.Close()
}