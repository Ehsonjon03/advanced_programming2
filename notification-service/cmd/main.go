package main

import (
	"context"
	"log"
	"notification-service/internal/provider" // создадим этот пакет
	"notification-service/internal/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// 1. Подключение к Redis (для Idempotency Check)
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Проверка подключения к Redis
	ctxRedis, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctxRedis).Err(); err != nil {
		log.Printf("Предупреждение: Redis не доступен, идемпотентность не гарантирована: %v", err)
	}

	// 2. Настройка Адаптера (Adapter Pattern)
	// В будущем здесь можно будет через if-else выбирать между Real и Simulated провайдером
	var emailProvider provider.EmailProvider
	emailProvider = &provider.SimulatedEmailProvider{}

	// 3. Создаем контекст для Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Передаем зависимости (Redis и Адаптер) в сервис
	// Тебе нужно будет обновить конструктор NewNotificationService
	svc := service.NewNotificationService(rdb, emailProvider)

	done := make(chan struct{})
	go func() {
		svc.Consume(ctx, rabbitURL)
		close(done)
	}()

	log.Println("Notification Worker запущен с поддержкой Retries и Idempotency...")

	<-ctx.Done()
	log.Println("Получен сигнал завершения, завершаем работу...")

	<-done
	if rdb != nil {
		rdb.Close()
	}
	log.Println("Notification Service полностью остановлен.")
}
