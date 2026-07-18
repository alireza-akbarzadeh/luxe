package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	InvoiceNumber  string     `gorm:"uniqueIndex;not null" json:"invoice_number"`
	OrderID        uint       `gorm:"not null;index" json:"order_id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	PaymentID      *uint      `gorm:"index" json:"payment_id,omitempty"`
	Subtotal       float64    `gorm:"type:decimal(10,2);not null;default:0" json:"subtotal"`
	TaxAmount      float64    `gorm:"type:decimal(10,2);not null;default:0" json:"tax_amount"`
	ShippingAmount float64    `gorm:"type:decimal(10,2);not null;default:0" json:"shipping_amount"`
	TotalAmount    float64    `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Currency       string     `gorm:"not null;default:'USD'" json:"currency"`
	Status         string     `gorm:"not null;default:'issued';index" json:"status"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	BillingName    string     `json:"billing_name,omitempty"`
	BillingEmail   string     `json:"billing_email,omitempty"`
	Notes          string     `json:"notes,omitempty"`

	Order   Order    `gorm:"foreignKey:OrderID" json:"-"`
	User    User     `gorm:"foreignKey:UserID" json:"-"`
	Payment *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
}
