package repository

import (
	"database/sql"
	"order-service/internal/domain"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Save(o domain.Order) error {
	_, err := r.db.Exec("INSERT INTO orders (id, customer_id, item_name, amount, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt)
	return err
}

func (r *OrderRepo) GetByID(id string) (*domain.Order, error) {
	var o domain.Order
	query := `SELECT id, status FROM orders WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&o.ID, &o.Status)
	return &o, err
}

// Написал функцию GetByAmountRange.(защищает от SQL-инъекций. Это самая важная часть безопасности на уровне данных.)
func (r *OrderRepo) GetByAmountRange(min, max int64) ([]domain.Order, error) {
	// SQL запрос для поиска в диапазоне
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at 
		FROM orders 
		WHERE amount >= $1 AND amount <= $2
	`

	rows, err := r.db.Query(query, min, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		// Сканируем данные из БД в структуру
		err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.ItemName,
			&o.Amount,
			&o.Status,
			&o.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	// Проверка на ошибки после цикла
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
