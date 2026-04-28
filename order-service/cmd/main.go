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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"order-service/internal/generated/order"
)

func main() {
	// Загружаем .env, если он есть (локально), в Docker переменные придут из compose
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		// Дефолт для локального запуска (Mac), в Docker подхватится значение из yaml
		connStr = "postgres://zhuanz:password@localhost:5432/payment_db?sslmode=disable"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	streamPort := os.Getenv("STREAM_PORT")
	if streamPort == "" {
		streamPort = "50052"
	}

	// 1. Подключение к БД с механизмом Retry (важно для Docker)
	var db *sql.DB
	var err error

	log.Println("Подключение к БД...")
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("Успешное подключение к БД!")
			break
		}

		log.Printf("Попытка %d: БД пока не готова, ждем 2 сек... (%v)", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("Критическая ошибка: не удалось подключиться к БД после 5 попыток:", err)
	}

	repo := repository.NewOrderRepo(db)
	uc := usecase.NewOrderUseCase(repo)

	hHTTP := httpHandler.NewOrderHandler(uc)
	hGRPC := grpcHandler.NewOrderStreamHandler(repo)

	// 2. Инициализация gRPC сервера
	grpcServer := grpc.NewServer()
	order.RegisterOrderServiceServer(grpcServer, hGRPC)

	// 3. Инициализация HTTP сервера
	r := gin.Default()

	// Группируем или оставляем как есть, главное чтобы совпадало с Postman
	r.POST("/orders", hHTTP.CreateOrder)
	r.GET("/orders", hHTTP.GetOrders)

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: r,
	}

	// --- GRACEFUL SHUTDOWN LOGIC ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запуск gRPC
	go func() {
		lis, err := net.Listen("tcp", ":"+streamPort)
		if err != nil {
			log.Fatalf("Ошибка порта для gRPC: %v", err)
		}
		log.Printf("Order gRPC Streaming Server запущен на :%s", streamPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC сервер остановлен: %v", err)
		}
	}()

	// Запуск HTTP
	go func() {
		log.Printf("Order HTTP Service запущен на порту :%s", httpPort)
		// Внутри контейнера слушаем 8080, снаружи в Postman стучим в 8081
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка HTTP сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	sig := <-quit
	log.Printf("Получен сигнал %v, начинаем остановку...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке HTTP: %v", err)
	}

	grpcServer.GracefulStop()
	db.Close()

	log.Println("Order Service полностью остановлен.")
}
