package main

import (
	"database/sql"
	"log"
	"net"
	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc_handler"
	"payment-service/internal/usecase"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"payment-service/pkg/payment"
)

func main() {
	// 1. Подключение к БД
	db, err := sql.Open("postgres", "user=zhuanz dbname=payment_db sslmode=disable")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// 2. Инициализация слоев
	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)
	// Используем наш новый gRPC Handler
	handler := grpc_handler.NewPaymentGRPCHandler(uc)

	// 3. Создаем TCP слушатель на порту 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка при прослушивании порта: %v", err)
	}

	// 4. Создаем gRPC сервер
	s := grpc.NewServer()
	payment.RegisterPaymentServiceServer(s, handler)

	log.Println("Payment gRPC Server запущен на порту :50051")

	// 5. Запуск
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
