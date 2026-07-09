package models

import (
	"time"

	"gorm.io/gorm"
)

// Return represents a customer return/refund request, driven by the "return" workflow.
type Return struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	OrderID         uint    `gorm:"not null;index" json:"order_id"`
	UserID          uint    `gorm:"not null;index" json:"user_id"`
	Reason          string  `json:"reason,omitempty"`
	Status          string  `gorm:"not null;default:'requested'" json:"status"`
	ReturnType      string  `gorm:"type:varchar(32);not null;default:'refund'" json:"return_type"`
	RefundAmount    float64 `gorm:"type:decimal(10,2);not null;default:0" json:"refund_amount"`
	AdminNotes      string  `json:"admin_notes,omitempty"`
	ExchangeNotes   string  `json:"exchange_notes,omitempty"`
	WorkflowStateID *uint   `gorm:"index" json:"workflow_state_id,omitempty"`

	Order         *Order         `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	User          *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	WorkflowState *WorkflowState `gorm:"foreignKey:WorkflowStateID" json:"workflow_state,omitempty"`
}
