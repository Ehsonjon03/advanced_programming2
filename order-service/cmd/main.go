package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"order-service/internal/repository"
	"order-service/internal/usecase"

	grpcHandler "order-service/internal/transport/grpc"
	httpHandler "order-service/internal/transport/http"

	"order-service/internal/generated/order"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}

	// Читаем настройки из .env
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "postgres://zhuanz:password@localhost:5432/payment_db?sslmode=disable"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080" // Используй этот порт в Postman!
	}

	streamPort := os.Getenv("STREAM_PORT")
	if streamPort == "" {
		streamPort = "50052"
	}

	// 1. Подключение к PostgreSQL с Retry
	var db *sql.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			log.Println("Успешное подключение к БД!")
			break
		}
		log.Printf("Попытка %d: БД не готова, ждем 2 сек...", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Критическая ошибка БД:", err)
	}

	// 2. Подключение к Redis
	var rdb *redis.Client
	for i := 0; i < 5; i++ {
		rdb = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		err = rdb.Ping(context.Background()).Err()
		if err == nil {
			log.Println("Успешное подключение к Redis!")
			break
		}
		log.Printf("Попытка %d: Redis не готов, ждем 2 сек...", i+1)
		time.Sleep(2 * time.Second)
	}

	// 3. Сборка слоев (Dependency Injection)
	repo := repository.NewOrderRepo(db, rdb)
	uc := usecase.NewOrderUseCase(repo)

	hHTTP := httpHandler.NewOrderHandler(uc)
	hGRPC := grpcHandler.NewOrderStreamHandler(repo)

	// 4. Настройка gRPC
	grpcServer := grpc.NewServer()
	order.RegisterOrderServiceServer(grpcServer, hGRPC)

	// 5. Настройка HTTP (Gin)
	r := gin.Default()
	r.POST("/orders", hHTTP.CreateOrder)
	r.GET("/orders", hHTTP.GetOrders)
	r.GET("/orders/:id", hHTTP.GetByID) // Теперь GetByID доступен!

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	// 6. Запуск и Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		lis, err := net.Listen("tcp", ":"+streamPort)
		if err != nil {
			log.Fatalf("Ошибка порта gRPC: %v", err)
		}
		log.Printf("gRPC сервер на :%s", streamPort)
		grpcServer.Serve(lis)
	}()

	go func() {
		log.Printf("HTTP сервер на :%s", httpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка HTTP сервера: %v", err)
		}
	}()

	<-quit
	log.Println("Остановка сервисов...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	grpcServer.GracefulStop()
	db.Close()
	if rdb != nil {
		rdb.Close()
	}
	log.Println("Order Service остановлен.")
}
