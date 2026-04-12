package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"strings" // Добавлено для очистки строк
	"time"

	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc_handler"
	"payment-service/internal/usecase"

	"payment-service/pkg/payment"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

// Интерцептор для логирования (Бонусные баллы: +10%)
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("Запрос: %s | Время: %v | Ошибка: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func main() {
	// 1. Загрузка конфигурации [cite: 35, 61]
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются стандартные настройки")
	}

	dbConn := strings.TrimSpace(os.Getenv("DB_URL"))
	if dbConn == "" {
		dbConn = "user=zhuanz dbname=payment_db sslmode=disable"
	}

	grpcPort := strings.TrimSpace(os.Getenv("GRPC_PORT"))
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// 2. Подключение к БД
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Инициализация Clean Architecture [cite: 34, 52]
	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := grpc_handler.NewPaymentGRPCHandler(uc)

	// 3. TCP слушатель [cite: 35, 36]
	// Форматируем строго как ":50051"
	address := ":" + grpcPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		// Если порт всё еще "кривой", эта ошибка покажет точно, что в переменной
		log.Fatalf("Ошибка при прослушивании порта [%s]: %v", address, err)
	}

	// 4. gRPC сервер с Интерцептором [cite: 72]
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	payment.RegisterPaymentServiceServer(s, handler)

	log.Printf("Payment gRPC Server запущен на порту %s", address)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
