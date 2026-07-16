package models

import "time"

// CollectionProduct links a product to a manual collection with explicit ordering.
type CollectionProduct struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CollectionID uint      `gorm:"not null;index" json:"collection_id"`
	ProductID    uint      `gorm:"not null;index" json:"product_id"`
	SortOrder    int       `gorm:"not null;default:0" json:"sort_order"`
	Position     int       `gorm:"not null;default:0" json:"position"`
	IsPinned     bool      `gorm:"not null;default:false" json:"is_pinned"`
	IsHidden     bool      `gorm:"not null;default:false" json:"is_hidden"`
	BoostScore   int       `gorm:"not null;default:0" json:"boost_score"`
	Product      *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (CollectionProduct) TableName() string {
	return "collection_products"
}
