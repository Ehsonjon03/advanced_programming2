package main

import (
	"context"
	"log"
	"notification-service/internal/service"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	// Создаем контекст, который отменится при сигнале прерывания
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	svc := service.NewNotificationService()

	// Запускаем Consume в отдельной горутине, чтобы main не блокировался навсегда
	done := make(chan struct{})
	go func() {
		svc.Consume(ctx, rabbitURL)
		close(done)
	}()

	log.Println("Notification Service запущен...")

	// Ждем сигнала завершения
	<-ctx.Done()
	log.Println("Получен сигнал завершения, завершаем работу...")

	// Ждем, пока горутина с Consume закончит обработку последнего сообщения
	<-done
	log.Println("Notification Service полностью остановлен.")
}
