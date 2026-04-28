package domain

type PaymentCompletedEvent struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"` // В учебных целях можно захардкодить
	Status        string `json:"status"`
}
