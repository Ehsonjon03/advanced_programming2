package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"order-service/internal/repository"
	transport "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	connStr := "user=zhuanz dbname=order_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к order_db:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("База order_db недоступна:", err)
	}
	fmt.Println("Успешно: Order Service подключен к order_db")

	repo := repository.NewOrderRepo(db)

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	uc := usecase.NewOrderUseCase(repo, httpClient)
	handler := transport.NewOrderHandler(uc)

	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)

	// Добавили строку r.GET("/orders", handler.GetOrders).
	// Теперь GET запрос на /orders будет вызывать нашу валидацию и поиск в БД
	r.GET("/orders", handler.GetOrders)

	// 4. ЗАПУСК на порту 8080
	fmt.Println("Order Service запущен на порту :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Не удалось запустить сервер:", err)
	}
}
