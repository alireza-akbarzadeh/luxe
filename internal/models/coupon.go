package models

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Code               string    `gorm:"uniqueIndex;not null" json:"code" validate:"required,min=3,max=50"`
	Description        string    `json:"description,omitempty"`
	DiscountType       string    `gorm:"not null" json:"discount_type" validate:"oneof=percentage fixed"`
	DiscountValue      float64   `gorm:"type:decimal(10,2);not null" json:"discount_value" validate:"gt=0"`
	MinimumOrderAmount float64   `gorm:"type:decimal(10,2);default:0" json:"minimum_order_amount"`
	MaxDiscountAmount  *float64  `gorm:"type:decimal(10,2)" json:"max_discount_amount,omitempty"`
	UsageLimit         int       `gorm:"default:1" json:"usage_limit" validate:"gt=0"`
	UsedCount          int       `gorm:"not null;default:0" json:"used_count"`
	StartDate          time.Time `gorm:"not null" json:"start_date"`
	EndDate            time.Time `gorm:"not null" json:"end_date"`
	IsActive           bool      `gorm:"default:true" json:"is_active"`
	WorkflowStateID    *uint     `gorm:"index" json:"workflow_state_id,omitempty"`
	WorkflowState      *WorkflowState `gorm:"foreignKey:WorkflowStateID;references:ID" json:"workflow_state,omitempty"`
}

type CouponUsage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	CouponID       uint      `gorm:"not null;index" json:"coupon_id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	OrderID        uint      `gorm:"not null;index" json:"order_id"`
	DiscountAmount float64   `gorm:"type:decimal(10,2);not null" json:"discount_amount"`

	// Associations
	Coupon Coupon `gorm:"foreignKey:CouponID" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`
	Order  Order  `gorm:"foreignKey:OrderID" json:"-"`
}
