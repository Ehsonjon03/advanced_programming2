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

// НОВЫЙ МЕТОД: Получение списка с фильтрацией
func (r *PaymentRepo) List(status string) ([]domain.Payment, error) {
	var payments []domain.Payment
	query := "SELECT id, order_id, transaction_id, amount, status FROM payments"
	var args []interface{}

	// Если статус передан (не пустой), добавляем фильтрацию WHERE
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
