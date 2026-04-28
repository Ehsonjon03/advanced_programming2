package usecase

import (
	"context"
	"payment-service/internal/domain"
	"payment-service/internal/repository"

	"github.com/google/uuid"
)

// MessageProducer — интерфейс для отправки событий.
// Это позволяет соблюсти "Separation of Concerns".
type MessageProducer interface {
	PublishPaymentEvent(ctx context.Context, event domain.PaymentCompletedEvent) error
}

type PaymentUseCase struct {
	repo     *repository.PaymentRepo
	producer MessageProducer // Добавляем продюсера в структуру
}

// Обновляем конструктор, чтобы он принимал продюсера
func NewPaymentUseCase(r *repository.PaymentRepo, p MessageProducer) *PaymentUseCase {
	return &PaymentUseCase{
		repo:     r,
		producer: p,
	}
}

// GetAll вызывает репозиторий для получения списка платежей
func (u *PaymentUseCase) GetAll(status string) ([]domain.Payment, error) {
	return u.repo.List(status)
}

// Authorize — логика создания нового платежа с отправкой события
func (u *PaymentUseCase) Authorize(orderID string, amount int64) (domain.Payment, error) {
	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	p := domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	// 1. Сохраняем в базу данных
	err := u.repo.Save(p)
	if err != nil {
		return p, err
	}

	// 2. Подготавливаем данные для уведомления
	// Здесь мы передаем order_id, amount, status и email (имитируем его пока)
	event := domain.PaymentCompletedEvent{
		OrderID:       p.OrderID,
		Amount:        p.Amount,
		CustomerEmail: "student-aitu@example.com", // Можно будет брать из параметров позже
		Status:        p.Status,
	}

	// 3. Публикуем событие в брокер сообщений [cite: 18, 31]
	// Используем context.Background(), так как это асинхронное действие
	_ = u.producer.PublishPaymentEvent(context.Background(), event)

	return p, nil
}
