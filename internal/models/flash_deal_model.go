package models

import "time"

// FlashDeal is a time-limited promotional offer for a product on the homepage.
type FlashDeal struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductID     uint      `gorm:"not null;index" json:"product_id"`
	Product       *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	EndsAt        time.Time `gorm:"not null;index" json:"ends_at"`
	QuantityLimit *int      `json:"quantity_limit,omitempty"`
	SortOrder     int       `gorm:"not null;default:0" json:"sort_order"`
	Status        string    `gorm:"size:32;not null;default:'active'" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (FlashDeal) TableName() string {
	return "flash_deals"
}
