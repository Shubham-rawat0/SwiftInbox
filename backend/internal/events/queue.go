package queue

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
    Conn *amqp.Connection
    Ch   *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
    conn, err := amqp.Dial(url)
    log.Println("[WORKER] connecting to rabbitmq")
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
    r.Ch.Close()
    r.Conn.Close()
}