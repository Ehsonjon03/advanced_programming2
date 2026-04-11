package usecase

import (
	"context"
	"log"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"order-service/pkg/payment" // Твои скопированные файлы
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderUseCase struct {
	repo *repository.OrderRepo
}

func NewOrderUseCase(r *repository.OrderRepo) *OrderUseCase {
	return &OrderUseCase{
		repo: r,
	}
}

func (u *OrderUseCase) CreateOrder(ord domain.Order) (string, error) {
	ord.Status = "Pending"
	ord.CreatedAt = time.Now()

	// 1. Сохраняем заказ в базу данных
	if err := u.repo.Save(ord); err != nil {
		return "", err
	}

	// 2. Устанавливаем соединение с Payment Service (порт 50051)
	// Используем insecure, так как у нас локальная разработка без SSL
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Не удалось подключиться к Payment Service: %v", err)
		return "Order created, but Payment Service is unreachable", nil
	}
	defer conn.Close()

	// 3. Создаем gRPC клиент
	client := payment.NewPaymentServiceClient(conn)

	// 4. Вызываем метод ProcessPayment (как обычную функцию!)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.ProcessPayment(ctx, &payment.PaymentRequest{
		OrderId: ord.ID,
		Amount:  ord.Amount,
	})

	if err != nil {
		log.Printf("Ошибка при вызове gRPC: %v", err)
		return "Order created, but payment failed", nil
	}

	// 5. Логируем результат от платежки
	log.Printf("Payment Status: %s, Transaction ID: %s", resp.Status, resp.TransactionId)

	return "Order created and processed via gRPC. Status: " + resp.Status, nil
}

func (u *OrderUseCase) GetFilteredOrders(min, max int64) ([]domain.Order, error) {
	return u.repo.GetByAmountRange(min, max)
}
