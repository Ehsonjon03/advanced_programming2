package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification-service/internal/provider" // Твой адаптер

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type NotificationService struct {
	rdb           *redis.Client
	emailProvider provider.EmailProvider
}

// Обновленный конструктор: теперь принимает зависимости
func NewNotificationService(rdb *redis.Client, emailProvider provider.EmailProvider) *NotificationService {
	return &NotificationService{
		rdb:           rdb,
		emailProvider: emailProvider,
	}
}

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

	for {
		select {
		case <-ctx.Done():
			log.Println("[Notification] Stopping consumer gracefully...")
			return

		case d, ok := <-msgs:
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

			// 1. ПРОВЕРКА IDEMPOTENCY (через Redis)
			// Если ключ удалось поставить (SetNX), значит сообщение новое
			cacheKey := fmt.Sprintf("processed_order:%s", event.OrderID)
			isNew, err := s.rdb.SetNX(ctx, cacheKey, "completed", 24*time.Hour).Result()
			if err != nil || !isNew {
				log.Printf("[!] Duplicate detected or Redis error for Order %s. Skipping...", event.OrderID)
				_ = d.Ack(false)
				continue
			}

			// 2. BACKGROUND JOB С RETRY LOGIC (Exponential Backoff)
			go func(orderID, email, status string, msg amqp.Delivery) {
				maxRetries := 3
				backoff := 2 * time.Second // Стартуем с 2 секунд

				success := false
				for i := 1; i <= maxRetries; i++ {
					body := fmt.Sprintf("Your payment for Order %s is %s", orderID, status)

					// Используем Адаптер (emailProvider)
					err := s.emailProvider.SendEmail(email, body)
					if err == nil {
						success = true
						log.Printf("[Notification] Email successfully sent to %s", email)
						break
					}

					log.Printf("[Retry %d/%d] Failed to send email to %s: %v. Retrying in %v...", i, maxRetries, email, err, backoff)
					time.Sleep(backoff)
					backoff *= 2 // Экспоненциально увеличиваем время: 2с -> 4с -> 8с
				}

				if success {
					_ = msg.Ack(false)
				} else {
					log.Printf("[CRITICAL] Could not send notification for Order %s after %d retries", orderID, maxRetries)
					// Если совсем не вышло, удаляем из Redis, чтобы можно было попробовать позже
					s.rdb.Del(ctx, cacheKey)
					_ = msg.Nack(false, false)
				}
			}(event.OrderID, event.Email, event.Status, d)
		}
	}
}
