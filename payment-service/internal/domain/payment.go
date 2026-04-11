package domain

type Payment struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"` // Сумма в центах [cite: 71]
	Status        string `json:"status"` // "Authorized", "Declined" [cite: 72]
}
