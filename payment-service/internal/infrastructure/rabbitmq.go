package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"payment-service/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQProducer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewRabbitMQProducer(url string) (*RabbitMQProducer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	q, err := ch.QueueDeclare(
		"payment.completed", // name
		true,                // durable (сохраняется при рестарте) [cite: 36]
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %v", err)
	}

	return &RabbitMQProducer{
		conn:    conn,
		channel: ch,
		queue:   q.Name,
	}, nil
}

func (p *RabbitMQProducer) PublishPaymentEvent(ctx context.Context, event domain.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx,
		"",      // exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Сообщение сохраняется на диске
		})
}

func (p *RabbitMQProducer) Close() {
	p.channel.Close()
	p.conn.Close()
}
