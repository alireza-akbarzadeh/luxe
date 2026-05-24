package dto

type PaymentRequest struct {
	OrderID  uint    `json:"order_id"`
	UserID   uint    `json:"user_id"`
	Amount   float64 `json:"amount"`
	Method   string  `json:"method"`
	Currency string  `json:"currency"`
}

type PaymentProviderResponse struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	RequiresCard bool   `json:"requires_card"`
}

type GetPaymentProviderQuery struct {
	IsActive *bool `form:"is_active"`
}
