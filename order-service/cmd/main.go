package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"order-service/internal/repository"
	"order-service/internal/usecase"

	// Правильные импорты для двух видов транспорта
	grpcHandler "order-service/internal/transport/grpc"
	httpHandler "order-service/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"order-service/internal/generated/order"
)

func main() {
	// Загружаем .env
	godotenv.Load()

	// Читаем настройки из .env
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "user=postgres dbname=order_db sslmode=disable"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	streamPort := os.Getenv("STREAM_PORT")
	if streamPort == "" {
		streamPort = "50052"
	}

	// 1. Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к order_db:", err)
	}

	// Инициализация слоев Clean Architecture [cite: 52, 53]
	repo := repository.NewOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo)

	// Хендлеры для разных протоколов
	hHTTP := httpHandler.NewOrderHandler(uc)
	hGRPC := grpcHandler.NewOrderStreamHandler(repo)

	// 2. ЗАПУСК gRPC СЕРВЕРА (Server-side Streaming) [cite: 39, 40]
	go func() {
		lis, err := net.Listen("tcp", ":"+streamPort)
		if err != nil {
			log.Fatalf("Ошибка порта для gRPC: %v", err)
		}

		s := grpc.NewServer()

		// Регистрация сервиса из сгенерированного кода [cite: 29, 42]
		order.RegisterOrderServiceServer(s, hGRPC)

		log.Printf("Order gRPC Streaming Server запущен на :%s", streamPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Ошибка запуска gRPC: %v", err)
		}
	}()

	// 3. ЗАПУСК HTTP СЕРВЕРА (Gin) [cite: 37]
	r := gin.Default()
	r.POST("/orders", hHTTP.CreateOrder)
	r.GET("/orders", hHTTP.GetOrders)

	fmt.Printf("Order HTTP Service запущен на порту :%s\n", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatal("Не удалось запустить HTTP сервер:", err)
	}
}
