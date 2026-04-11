package usecase

import (
	"bytes"
	"encoding/json"
	"net/http"
	"order-service/internal/domain"
	"order-service/internal/repository"
	"time"
)

type OrderUseCase struct {
	repo       *repository.OrderRepo
	httpClient *http.Client
}

func NewOrderUseCase(r *repository.OrderRepo, client *http.Client) *OrderUseCase {
	return &OrderUseCase{
		repo:       r,
		httpClient: client,
	}
}

func (u *OrderUseCase) CreateOrder(ord domain.Order) (string, error) {
	ord.Status = "Pending"
	ord.CreatedAt = time.Now()

	if err := u.repo.Save(ord); err != nil {
		return "", err
	}

	paymentReq, _ := json.Marshal(map[string]interface{}{
		"order_id": ord.ID,
		"amount":   ord.Amount,
	})

	resp, err := u.httpClient.Post("http://localhost:8081/payments", "application/json", bytes.NewBuffer(paymentReq))
	if err != nil {
		return "Failed: Payment Service timeout or error", nil
	}
	defer resp.Body.Close()

	return "Order created and processed", nil
}

// Добавил метод GetFilteredOrders.(Сейчас он просто вызывает репозиторий.)

func (u *OrderUseCase) GetFilteredOrders(min, max int64) ([]domain.Order, error) {
	
	return u.repo.GetByAmountRange(min, max)
}
