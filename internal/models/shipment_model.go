package models

import (
	"time"

	"gorm.io/gorm"
)

type Shipment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	OrderID           uint       `gorm:"not null;index" json:"order_id"`
	UserID            uint       `gorm:"not null;index" json:"user_id"`
	Carrier           string     `gorm:"not null" json:"carrier"`
	TrackingNumber    string     `json:"tracking_number,omitempty"`
	Status            string     `gorm:"not null;default:'pending'" json:"status"`
	WorkflowStateID   *uint      `gorm:"index" json:"workflow_state_id,omitempty"`
	ShippedAt         *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`

	// Shipping address
	AddressLine1 string `gorm:"not null" json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `gorm:"not null" json:"city"`
	State        string `json:"state,omitempty"`
	PostalCode   string `gorm:"not null" json:"postal_code"`
	Country      string `gorm:"not null" json:"country"`

	ProviderID    *uint   `gorm:"index" json:"provider_id,omitempty"`
	ShippingPrice float64 `gorm:"type:decimal(10,2);not null;default:0" json:"shipping_price"`

	Order    Order              `gorm:"foreignKey:OrderID" json:"-"`
	User     User               `gorm:"foreignKey:UserID" json:"-"`
	Provider *ShippingProviders `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

type ShippingProviders struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null;default:0" json:"price"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
}
