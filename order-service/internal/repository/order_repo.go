package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"order-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderRepo struct {
	db  *sql.DB
	rdb *redis.Client // Добавляем Redis клиент [cite: 48]
}

// Обновляем конструктор, чтобы принимать Redis клиент [cite: 58]
func NewOrderRepo(db *sql.DB, rdb *redis.Client) *OrderRepo {
	return &OrderRepo{
		db:  db,
		rdb: rdb,
	}
}

func (r *OrderRepo) Save(ctx context.Context, o domain.Order) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO orders (id, customer_id, item_name, amount, status, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, o.CreatedAt)

	// При создании заказа кэш можно не трогать, но если хочешь быть уверенным — удали ключ
	if err == nil {
		r.rdb.Del(ctx, "order:"+o.ID)
	}
	return err
}

// Реализация паттерна Cache-aside [cite: 20, 27]
func (r *OrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	cacheKey := "order:" + id

	// 1. Read Path: Сначала проверяем Redis
	val, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var o domain.Order
		if err := json.Unmarshal([]byte(val), &o); err == nil {
			log.Printf("Кэш найден для заказа: %s", id)
			return &o, nil // Cache Hit
		}
	}

	// 2. Если в кэше нет — идем в БД
	var o domain.Order
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`
	err = r.db.QueryRowContext(ctx, query, id).Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем в Redis с TTL (5 минут) [cite: 30, 60]
	data, _ := json.Marshal(o)
	r.rdb.Set(ctx, cacheKey, data, 5*time.Minute)

	return &o, nil
}

// Добавляем метод для обновления статуса (важен для инвалидации кэша)
func (r *OrderRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	// 1. Обновляем БД
	_, err := r.db.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return err
	}

	// 2. Invalidation: Удаляем ключ из Redis, чтобы не отдавать старые данные [cite: 31, 55]
	err = r.rdb.Del(ctx, "order:"+id).Err()
	if err != nil {
		log.Printf("Ошибка удаления кэша: %v", err)
	}

	return nil
}

func (r *OrderRepo) GetByAmountRange(ctx context.Context, min, max int64) ([]domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE amount >= $1 AND amount <= $2`

	rows, err := r.db.QueryContext(ctx, query, min, max)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
