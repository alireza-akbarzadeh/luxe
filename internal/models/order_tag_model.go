package models

import "time"

// OrderTag is an admin label attached to an order for filtering and triage.
type OrderTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	Tag       string    `gorm:"not null;index" json:"tag"`
}

func (OrderTag) TableName() string {
	return "order_tags"
}
