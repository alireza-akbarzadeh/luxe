package models

import (
	"time"

	"gorm.io/datatypes"
)

// InventoryAdjustment is an append-only ledger row for product stock changes.
type InventoryAdjustment struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProductID      uint           `gorm:"not null;index:idx_inventory_adjustments_product_created,priority:1" json:"product_id"`
	QuantityDelta  int            `gorm:"not null" json:"quantity_delta"`
	QuantityBefore int            `gorm:"not null" json:"quantity_before"`
	QuantityAfter  int            `gorm:"not null" json:"quantity_after"`
	AdjustmentType string         `gorm:"type:varchar(32);not null" json:"adjustment_type"`
	ReferenceType  string         `gorm:"type:varchar(32)" json:"reference_type,omitempty"`
	ReferenceID    *uint          `json:"reference_id,omitempty"`
	ActorUserID    *uint          `json:"actor_user_id,omitempty"`
	Note           string         `gorm:"not null;default:''" json:"note"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"index:idx_inventory_adjustments_product_created,priority:2,sort:desc" json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Actor   *User    `gorm:"foreignKey:ActorUserID" json:"actor,omitempty"`
}

func (InventoryAdjustment) TableName() string {
	return "inventory_adjustments"
}
