package usecase

import (
	"payment-service/internal/domain"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo *repository.PaymentRepo
}

func NewPaymentUseCase(r *repository.PaymentRepo) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
}

// GetAll вызывает репозиторий для получения списка платежей
func (u *PaymentUseCase) GetAll(status string) ([]domain.Payment, error) {
	return u.repo.List(status)
}

// Authorize — логика создания нового платежа
func (u *PaymentUseCase) Authorize(orderID string, amount int64) (domain.Payment, error) {
	status := "Authorized"
	if amount > 100000 { // Правило: > 100000 = Decline
		status = "Declined"
	}

	p := domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	return p, u.repo.Save(p)
}
