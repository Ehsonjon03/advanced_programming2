package usecase

import (
	"github.com/google/uuid"
	"payment-service/internal/domain"
	"payment-service/internal/repository"
)

type PaymentUseCase struct {
	repo *repository.PaymentRepo
}

func NewPaymentUseCase(r *repository.PaymentRepo) *PaymentUseCase {
	return &PaymentUseCase{repo: r}
}

func (u *PaymentUseCase) Authorize(orderID string, amount int64) (domain.Payment, error) {
	status := "Authorized"
	if amount > 100000 { // Правило: > 1000 единиц = Decline
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
