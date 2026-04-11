package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"order-service/internal/repository"
	transport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// Загружаем .env
	godotenv.Load()

	// Читаем настройки
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "user=zhuanz dbname=order_db sslmode=disable"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	// Адрес, где Order Service будет слушать gRPC стриминг
	streamPort := os.Getenv("STREAM_PORT")
	if streamPort == "" {
		streamPort = "50052"
	}

	// 1. Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к order_db:", err)
	}

	repo := repository.NewOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo)
	handler := transport.NewOrderHandler(uc)

	// 2. ЗАПУСК gRPC СЕРВЕРА (для стриминга) в отдельной горутине
	go func() {
		lis, err := net.Listen("tcp", ":"+streamPort)
		if err != nil {
			log.Fatalf("gRPC Stream: ошибка порта: %v", err)
		}
		s := grpc.NewServer()

		// Тут нужно будет зарегистрировать твой OrderTrackingServiceServer
		// payment.RegisterOrderTrackingServiceServer(s, &YourStreamingHandler{repo: repo})

		log.Printf("Order gRPC Streaming Server запущен на :%s", streamPort)
		s.Serve(lis)
	}()

	// 3. Настройка маршрутов (Gin HTTP)
	r := gin.Default()
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders", handler.GetOrders)

	fmt.Printf("Order HTTP Service запущен на порту :%s\n", httpPort)
	if err := r.Run(":" + httpPort); err != nil {
		log.Fatal("Не удалось запустить сервер:", err)
	}
}
