package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc_handler"
	"payment-service/internal/usecase"

	"github.com/joho/godotenv" // Не забудь сделать go get github.com/joho/godotenv
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"payment-service/pkg/payment"
)

// Интерцептор для логирования (Бонусные баллы)
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("Запрос: %s | Время: %v | Ошибка: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются стандартные настройки")
	}

	// Читаем настройки из .env
	dbConn := os.Getenv("DB_URL")
	if dbConn == "" {
		dbConn = "user=zhuanz dbname=payment_db sslmode=disable"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// 1. Подключение к БД
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := grpc_handler.NewPaymentGRPCHandler(uc)

	// 2. TCP слушатель
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Ошибка при прослушивании порта: %v", err)
	}

	// 3. Создаем gRPC сервер с интерцептором
	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)
	payment.RegisterPaymentServiceServer(s, handler)

	log.Printf("Payment gRPC Server запущен на порту :%s", grpcPort)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
