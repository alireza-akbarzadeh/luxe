package dto

type PaymentRequest struct {
	OrderID  uint    `json:"order_id"`
	UserID   uint    `json:"user_id"`
	Amount   float64 `json:"amount"`
	Method   string  `json:"method"`
	Currency string  `json:"currency"`
}
