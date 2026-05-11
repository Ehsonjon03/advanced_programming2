package usecase

import (
	"context"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"order-service/pkg/payment"
	"os"
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

// ДОБАВЛЕНО: Теперь этот метод существует и доступен для Handler
func (u *OrderUseCase) GetByID(id string) (*domain.Order, error) {
	return u.repo.GetByID(context.Background(), id)
}

func (u *OrderUseCase) CreateOrder(ord domain.Order) (string, error) {
	ord.Status = "Pending"
	ord.CreatedAt = time.Now()
	ctx := context.Background()

	if err := u.repo.Save(ctx, ord); err != nil {
		return "", err
	}

	paymentAddr := os.Getenv("PAYMENT_SERVICE_ADDR")
	if paymentAddr == "" {
		paymentAddr = "localhost:50051"
	}

	conn, err := grpc.Dial(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "Order created, but Payment Service is unreachable", nil
	}
	defer conn.Close()

	client := payment.NewPaymentServiceClient(conn)
	gRPCCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	resp, err := client.ProcessPayment(gRPCCtx, &payment.PaymentRequest{
		OrderId: ord.ID,
		Amount:  ord.Amount,
	})

	if err != nil {
		return "Order created, but payment failed", nil
	}

	// Обновляем статус после gRPC ответа
	_ = u.repo.UpdateStatus(ctx, ord.ID, resp.Status)

	return "Order created and processed via gRPC. Status: " + resp.Status, nil
}

func (u *OrderUseCase) GetFilteredOrders(min, max int64) ([]domain.Order, error) {
	return u.repo.GetByAmountRange(context.Background(), min, max)
}
