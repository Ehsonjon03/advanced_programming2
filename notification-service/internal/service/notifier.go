package service

import (
	"context" // Добавлен контекст
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationService struct {
	processedOrders map[string]bool
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		processedOrders: make(map[string]bool),
	}
}

// Добавляем ctx в параметры метода
func (s *NotificationService) Consume(ctx context.Context, url string) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, _ := ch.QueueDeclare("payment.completed", true, false, false, false, nil)

	_ = ch.Qos(1, 0, false)

	msgs, _ := ch.Consume(q.Name, "", false, false, false, false, nil)

	log.Println("[*] Waiting for payment events. To exit press CTRL+C")

	// Используем бесконечный цикл с select для возможности прерывания
	for {
		select {
		case <-ctx.Done(): // Сигнал на остановку из main.go
			log.Println("[Notification] Stopping consumer gracefully...")
			return

		case d, ok := <-msgs: // Получение сообщения из очереди
			if !ok {
				log.Println("[Notification] Message channel closed")
				return
			}

			var event struct {
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
				Email   string `json:"customer_email"`
			}

			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error unmarshalling: %s", err)
				_ = d.Nack(false, false)
				continue
			}

			// ПРОВЕРКА IDEMPOTENCY
			if s.processedOrders[event.OrderID] {
				log.Printf("[!] Duplicate detected for Order %s. Skipping...", event.OrderID)
				_ = d.Ack(false)
				continue
			}

			// Имитация отправки письма
			log.Printf("[Notification] Email sent to %s: Your payment for Order %s is %s", event.Email, event.OrderID, event.Status)

			s.processedOrders[event.OrderID] = true

			// MANUAL ACK
			_ = d.Ack(false)
		}
	}
}
