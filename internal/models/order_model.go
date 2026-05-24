package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	UserID      uint    `gorm:"not null;index" json:"user_id"`
	OrderNumber string  `gorm:"uniqueIndex;not null" json:"order_number"`
	Status      string  `gorm:"not null;default:'pending'" json:"status"`
	TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Currency    string  `gorm:"not null;default:'USD'" json:"currency"`
	Notes       string  `json:"notes,omitempty"`

	// Address IDs (separate addresses table could be added later)
	BillingAddressID  *uint `json:"billing_address_id,omitempty"`
	ShippingAddressID *uint `json:"shipping_address_id,omitempty"`

	PaymentID  *uint `json:"payment_id,omitempty"`
	ShipmentID *uint `json:"shipment_id,omitempty"`

	// Associations
	User  User        `gorm:"foreignKey:UserID" json:"-"`
	Items []OrderItem `json:"items,omitempty"`

	Payment  *Payment  `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
	Shipment *Shipment `gorm:"foreignKey:OrderID" json:"shipment,omitempty"`
}

type OrderItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	OrderID   uint    `gorm:"not null;index" json:"order_id"`
	ProductID uint    `gorm:"not null;index" json:"product_id"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	Total     float64 `gorm:"->;type:decimal(10,2)" json:"total"` // read-only

	Order   Order   `gorm:"foreignKey:OrderID" json:"-"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}
