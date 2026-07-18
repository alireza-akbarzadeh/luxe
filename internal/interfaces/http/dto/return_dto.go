package dto

import "time"

const (
	ReturnTypeRefund   = "refund"
	ReturnTypeExchange = "exchange"
)

// CreateReturnRequest is submitted by a customer to open a return/refund case.
type CreateReturnRequest struct {
	OrderID       uint   `json:"order_id" validate:"required,gt=0"`
	Reason        string `json:"reason" validate:"required,min=3,max=512"`
	ReturnType    string `json:"return_type" validate:"omitempty,oneof=refund exchange"`
	ExchangeNotes string `json:"exchange_notes" validate:"omitempty,max=512"`
}

// ReturnResponse is the API view of a return request.
type ReturnResponse struct {
	ID              uint       `json:"id"`
	OrderID         uint       `json:"order_id"`
	UserID          uint       `json:"user_id"`
	Reason          string     `json:"reason,omitempty"`
	Status          string     `json:"status"`
	ReturnType      string     `json:"return_type"`
	RefundAmount    float64    `json:"refund_amount"`
	AdminNotes      string     `json:"admin_notes,omitempty"`
	ExchangeNotes   string     `json:"exchange_notes,omitempty"`
	WorkflowStateID *uint      `json:"workflow_state_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	OrderNumber     string     `json:"order_number,omitempty"`
	CustomerName    string     `json:"customer_name,omitempty"`
	CustomerEmail   string     `json:"customer_email,omitempty"`
	State           *StateView `json:"state,omitempty"`
}

// AdminReturnListFilters supports admin listing of return requests.
type AdminReturnListFilters struct {
	Status        string `form:"status"`
	ReturnType    string `form:"return_type"`
	WorkflowState string `form:"workflow_state"`
	Search        string `form:"search"`
	UserID        *uint  `form:"user_id"`
	Limit         int    `form:"limit"`
	Offset        int    `form:"offset"`
}

// UpdateReturnNotesRequest updates admin-only notes on a return.
type UpdateReturnNotesRequest struct {
	AdminNotes string `json:"admin_notes" validate:"max=2048"`
}

// PerformReturnTransitionRequest is an admin action on a return workflow.
type PerformReturnTransitionRequest struct {
	Event string `json:"event" validate:"required,min=1,max=64"`
	Note  string `json:"note"  validate:"omitempty,max=512"`
}

// AdminReturnStats aggregates return analytics for the admin dashboard.
type AdminReturnStats struct {
	Total       int64              `json:"total"`
	Open        int64              `json:"open"`
	RefundTotal float64            `json:"refund_total"`
	ByStatus    map[string]int64   `json:"by_status"`
	ByType      map[string]int64   `json:"by_type"`
	Last7Days   []ReturnDailyCount `json:"last_7_days"`
}

// ReturnDailyCount is a single day bucket for return volume charts.
type ReturnDailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}
