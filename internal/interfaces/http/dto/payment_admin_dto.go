package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// AdminPaymentListFilters filters the admin payments list.
type AdminPaymentListFilters struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	Method   string `form:"method"`
	UserID   *uint  `form:"user_id"`
	OrderID  *uint  `form:"order_id"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

// AdminPaymentListItem is a row in the admin payments table.
type AdminPaymentListItem struct {
	ID              uint      `json:"id"`
	OrderID         uint      `json:"order_id"`
	OrderNumber     string    `json:"order_number,omitempty"`
	UserID          uint      `json:"user_id"`
	CustomerName    string    `json:"customer_name,omitempty"`
	CustomerEmail   string    `json:"customer_email,omitempty"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Method          string    `json:"method"`
	Status          string    `json:"status"`
	TransactionID   string    `json:"transaction_id,omitempty"`
	StripeSessionID string    `json:"stripe_session_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AdminPaymentListData wraps the paginated admin payments response.
type AdminPaymentListData struct {
	Payments []AdminPaymentListItem `json:"payments"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	Limit    int                    `json:"limit"`
}

// AdminPaymentDetailResponse powers the admin payment detail view.
type AdminPaymentDetailResponse struct {
	ID              uint                   `json:"id"`
	OrderID         uint                   `json:"order_id"`
	OrderNumber     string                 `json:"order_number,omitempty"`
	UserID          uint                   `json:"user_id"`
	CustomerName    string                 `json:"customer_name,omitempty"`
	CustomerEmail   string                 `json:"customer_email,omitempty"`
	Amount          float64                `json:"amount"`
	Currency        string                 `json:"currency"`
	Method          string                 `json:"method"`
	Status          string                 `json:"status"`
	TransactionID   string                 `json:"transaction_id,omitempty"`
	StripeSessionID string                 `json:"stripe_session_id,omitempty"`
	GatewayResponse map[string]interface{} `json:"gateway_response,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// PaymentsSummaryResponse holds admin payments KPI counters.
type PaymentsSummaryResponse struct {
	TotalCount     int64   `json:"total_count"`
	CompletedCount int64   `json:"completed_count"`
	PendingCount   int64   `json:"pending_count"`
	FailedCount    int64   `json:"failed_count"`
	RefundedCount  int64   `json:"refunded_count"`
	TotalVolume    float64 `json:"total_volume"`
}

// ToAdminPaymentListItem maps a preloaded payment to a list row.
func ToAdminPaymentListItem(p *models.Payment) AdminPaymentListItem {
	item := AdminPaymentListItem{
		ID:              p.ID,
		OrderID:         p.OrderID,
		UserID:          p.UserID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		Method:          p.Method,
		Status:          p.Status,
		TransactionID:   p.TransactionID,
		StripeSessionID: p.StripeSessionID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
	if p.Order.ID != 0 {
		item.OrderNumber = p.Order.OrderNumber
	}
	if p.User.ID != 0 {
		item.CustomerName = strings.TrimSpace(p.User.FirstName + " " + p.User.LastName)
		item.CustomerEmail = p.User.Email
	}
	return item
}

// ToAdminPaymentDetail maps a fully preloaded payment to the admin detail response.
func ToAdminPaymentDetail(p *models.Payment) AdminPaymentDetailResponse {
	resp := AdminPaymentDetailResponse{
		ID:              p.ID,
		OrderID:         p.OrderID,
		UserID:          p.UserID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		Method:          p.Method,
		Status:          p.Status,
		TransactionID:   p.TransactionID,
		StripeSessionID: p.StripeSessionID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
	if p.Order.ID != 0 {
		resp.OrderNumber = p.Order.OrderNumber
	}
	if p.User.ID != 0 {
		resp.CustomerName = strings.TrimSpace(p.User.FirstName + " " + p.User.LastName)
		resp.CustomerEmail = p.User.Email
	}
	if len(p.GatewayResponse) > 0 {
		var parsed map[string]interface{}
		if err := json.Unmarshal(p.GatewayResponse, &parsed); err == nil {
			resp.GatewayResponse = parsed
		}
	}
	return resp
}
