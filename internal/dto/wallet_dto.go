package dto

import "time"

type DepositRequest struct {
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod string  `json:"payment_method" validate:"omitempty,oneof=stripe mock"`
}

type DepositResponse struct {
	TransactionID   uint   `json:"transaction_id"`
	Status          string `json:"status"`
	CheckoutURL     string `json:"checkout_url,omitempty"`
	StripeSessionID string `json:"stripe_session_id,omitempty"`
}

type AdminAdjustRequest struct {
	UserID      uint    `json:"user_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	Description string  `json:"description" validate:"required"`
}

type WalletListFilters struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

type WalletResponse struct {
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

type TransactionResponse struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	ReferenceType string    `json:"reference_type,omitempty"`
	ReferenceID   *uint     `json:"reference_id,omitempty"`
	Description   string    `json:"description,omitempty"`
	BalanceAfter  float64   `json:"balance_after"`
	Status        string    `json:"status"`
}

type WalletDetailResponse struct {
	Balance      float64               `json:"balance"`
	Currency     string                `json:"currency"`
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type WithdrawRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description,omitempty"`
}
