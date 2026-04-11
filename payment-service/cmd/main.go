package main

import (
	"database/sql"
	"fmt"
	"log"
	"payment-service/internal/repository"
	"payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // Драйвер для Postgres
)

func main() {
	// Твой существующий код подключения к БД...
	db, err := sql.Open("postgres", "user=zhuanz dbname=payment_db sslmode=disable")

	// Собираем слои Чистой Архитектуры
	repo := repository.NewPaymentRepo(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := http.NewPaymentHandler(uc)

	r := gin.Default()

	if err != nil {
		log.Fatal(err)
	}

	// Настраиваем эндпоинты согласно заданию [cite: 88, 91]
	r.POST("/payments", handler.CreatePayment)

	// Запускаем сервис
	fmt.Println("Payment Service запущен на порту :8081")
	r.Run(":8081")
}
