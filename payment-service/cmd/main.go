package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"payment-service/internal/infrastructure"
	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc_handler"
	"payment-service/internal/usecase"

	"payment-service/pkg/payment"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("Запрос: %s | Время: %v | Ошибка: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения")
	}

	dbConn := strings.TrimSpace(os.Getenv("DB_URL"))
	if dbConn == "" {
		// Замени на строку подключения для Docker
		// Было: user=zhuanz dbname=payment_db
		dbConn = "postgres://zhuanz:password@postgres:5432/payment_db?sslmode=disable"
	}

	grpcPort := strings.TrimSpace(os.Getenv("GRPC_PORT"))
	if grpcPort == "" {
		grpcPort = "50051"
	}

	rabbitURL := os.Getenv("RABBIT_URL") // Было RABBITMQ_URL в коде
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	// 1. Подключение к БД
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// 2. Инициализация RabbitMQ Producer
	producer, err := infrastructure.NewRabbitMQProducer(rabbitURL)
	if err != nil {
		log.Printf("Предупреждение: не удалось подключиться к RabbitMQ: %v", err)
	}

	// 3. Инициализация слоев
	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo, producer)
	handler := grpc_handler.NewPaymentGRPCHandler(uc)

	address := ":" + grpcPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Ошибка при прослушивании порта [%s]: %v", address, err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	payment.RegisterPaymentServiceServer(s, handler)

	// --- GRACEFUL SHUTDOWN LOGIC START ---

	// Канал для прослушивания сигналов прерывания от ОС
	quit := make(chan os.Signal, 1)
	// SIGINT - это Ctrl+C, SIGTERM - сигнал на завершение от Docker/Kubernetes
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине, чтобы он не блокировал основной поток
	go func() {
		log.Printf("Payment gRPC Server запущен на порту %s", address)
		if err := s.Serve(lis); err != nil {
			log.Printf("Сервер остановлен: %v", err)
		}
	}()

	// Программа "замирает" здесь, пока не получит сигнал из канала quit
	sig := <-quit
	log.Printf("Получен сигнал завершения (%v). Начинаем Graceful Shutdown...", sig)

	// Даем серверу 5 секунд на то, чтобы завершить текущие запросы
	s.GracefulStop()
	log.Println("gRPC сервер успешно остановлен.")

	// Закрываем соединение с RabbitMQ
	if producer != nil {
		producer.Close()
		log.Println("Соединение с RabbitMQ закрыто.")
	}

	// Закрываем соединение с БД
	if err := db.Close(); err != nil {
		log.Printf("Ошибка при закрытии БД: %v", err)
	} else {
		log.Println("Соединение с БД закрыто.")
	}

	log.Println("Payment Service полностью остановлен.")
}
