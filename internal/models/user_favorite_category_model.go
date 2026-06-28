package models

import "time"

// UserFavoriteCategory stores a shopper's preferred category for homepage personalization.
type UserFavoriteCategory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;uniqueIndex:idx_user_favorite_category_pair" json:"user_id"`
	CategoryID uint      `gorm:"not null;uniqueIndex:idx_user_favorite_category_pair" json:"category_id"`
	Category   *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UserFavoriteCategory) TableName() string {
	return "user_favorite_categories"
}
