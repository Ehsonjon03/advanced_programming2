package main

import (
	"database/sql"
	"fmt"
	"log"
	"order-service/internal/repository"
	transport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Подключение к базе данных
	connStr := "user=zhuanz dbname=order_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к order_db:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("База order_db недоступна:", err)
	}
	fmt.Println("Успешно: Order Service подключен к order_db")

	// 2. Инициализация слоев (Чистая Архитектура)
	repo := repository.NewOrderRepo(db)

	// Теперь httpClient здесь не нужен, так как UseCase использует gRPC
	uc := usecase.NewOrderUseCase(repo)
	handler := transport.NewOrderHandler(uc)

	// 3. Настройка маршрутов (Gin)
	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders", handler.GetOrders)

	// 4. Запуск сервера на порту 8080
	fmt.Println("Order Service запущен на порту :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Не удалось запустить сервер:", err)
	}
}
