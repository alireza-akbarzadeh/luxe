package models

import (
	"time"

	"gorm.io/gorm"
)

// ReverseMarketplaceRequest is a buyer-posted wanted listing.
type ReverseMarketplaceRequest struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"not null;default:''" json:"description"`
	Category    string         `gorm:"not null;default:''" json:"category"`
	BudgetMin   *float64       `gorm:"type:decimal(10,2)" json:"budget_min,omitempty"`
	BudgetMax   *float64       `gorm:"type:decimal(10,2)" json:"budget_max,omitempty"`
	Status      string         `gorm:"not null;default:'open';index" json:"status"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Offers      []ReverseMarketplaceOffer `gorm:"foreignKey:RequestID" json:"offers,omitempty"`
}

func (ReverseMarketplaceRequest) TableName() string {
	return "reverse_marketplace_requests"
}

// ReverseMarketplaceOffer is a vendor response to a buyer request.
type ReverseMarketplaceOffer struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	RequestID     uint           `gorm:"not null;index;uniqueIndex:idx_reverse_offer_request_store" json:"request_id"`
	StoreID       uint           `gorm:"not null;index;uniqueIndex:idx_reverse_offer_request_store" json:"store_id"`
	VendorUserID  uint           `gorm:"not null;index" json:"vendor_user_id"`
	Message       string         `gorm:"not null;default:''" json:"message"`
	OfferedPrice  float64        `gorm:"type:decimal(10,2);not null" json:"offered_price"`
	Status        string         `gorm:"not null;default:'pending'" json:"status"`
	Request       *ReverseMarketplaceRequest `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	Store         *Store         `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

func (ReverseMarketplaceOffer) TableName() string {
	return "reverse_marketplace_offers"
}
