package dto

import "time"

// CreateReturnRequest is submitted by a customer to open a return/refund case.
type CreateReturnRequest struct {
	OrderID uint   `json:"order_id" validate:"required,gt=0"`
	Reason  string `json:"reason"   validate:"required,min=3,max=512"`
}

// ReturnResponse is the API view of a return request.
type ReturnResponse struct {
	ID              uint       `json:"id"`
	OrderID         uint       `json:"order_id"`
	UserID          uint       `json:"user_id"`
	Reason          string     `json:"reason,omitempty"`
	Status          string     `json:"status"`
	RefundAmount    float64    `json:"refund_amount"`
	WorkflowStateID *uint      `json:"workflow_state_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	OrderNumber     string     `json:"order_number,omitempty"`
	State           *StateView `json:"state,omitempty"`
}

// AdminReturnListFilters supports admin listing of return requests.
type AdminReturnListFilters struct {
	Status string `form:"status"`
	UserID *uint  `form:"user_id"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// PerformReturnTransitionRequest is an admin action on a return workflow.
type PerformReturnTransitionRequest struct {
	Event string `json:"event" validate:"required,min=1,max=64"`
	Note  string `json:"note"  validate:"omitempty,max=512"`
}
