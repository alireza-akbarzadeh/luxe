package models

import (
	"time"

	"gorm.io/gorm"
)

// ShopLook is a shoppable lifestyle image with tagged product hotspots.
type ShopLook struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"not null;default:''" json:"description"`
	ImageURL    string         `gorm:"column:image_url;not null;default:''" json:"image_url"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	Tags        []ShopLookTag  `gorm:"foreignKey:ShopLookID" json:"tags,omitempty"`
}

func (ShopLook) TableName() string {
	return "shop_looks"
}

// ShopLookTag pins a product on a shop look image (percent-based coordinates).
type ShopLookTag struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ShopLookID uint      `gorm:"not null;index" json:"shop_look_id"`
	ProductID  uint      `gorm:"not null;index" json:"product_id"`
	XPercent   float64   `gorm:"type:numeric(5,2);not null" json:"x_percent"`
	YPercent   float64   `gorm:"type:numeric(5,2);not null" json:"y_percent"`
	Label      string    `gorm:"not null;default:''" json:"label"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
	Product    *Product  `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

func (ShopLookTag) TableName() string {
	return "shop_look_tags"
}
