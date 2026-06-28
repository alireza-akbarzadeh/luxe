package models

import "time"

// UserProductView tracks recently viewed products for continue-shopping sections.
type UserProductView struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_product_view_pair" json:"user_id"`
	ProductID uint      `gorm:"not null;uniqueIndex:idx_user_product_view_pair" json:"product_id"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ViewedAt  time.Time `gorm:"not null;index" json:"viewed_at"`
}

func (UserProductView) TableName() string {
	return "user_product_views"
}
