package domain

import "time"

type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ItemName   string    `json:"item_name"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// Добавил интерфейс OrderRepository.(уметь фильтровать заказы)

type OrderRepository interface {
	GetByAmountRange(min, max int64) ([]Order, error)
}
