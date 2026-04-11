package repository

import (
	"database/sql"
	"payment-service/internal/domain"
)

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Save(p domain.Payment) error {
	_, err := r.db.Exec("INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1, $2, $3, $4, $5)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status)
	return err
}
