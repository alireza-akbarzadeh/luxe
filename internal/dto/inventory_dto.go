package dto

import "time"

type ListInventoryRequest struct {
	Page        int    `form:"page,default=1"`
	Limit       int    `form:"limit,default=20"`
	Search      string `form:"search"`
	StockStatus string `form:"stock_status"`
	Sort        string `form:"sort"`
}

type AdjustInventoryRequest struct {
	ProductID uint   `json:"product_id" validate:"required,gt=0"`
	Delta     int    `json:"delta" validate:"required"`
	Reason    string `json:"reason" validate:"required,oneof=receive correction damage shrinkage cycle_count other"`
	Note      string `json:"note" validate:"omitempty,max=500"`
}

type BulkInventoryAdjustRow struct {
	SKU   string `json:"sku" validate:"required,min=1,max=64"`
	Delta int    `json:"delta" validate:"required"`
	Note  string `json:"note" validate:"omitempty,max=200"`
}

type BulkAdjustInventoryRequest struct {
	Reason string                   `json:"reason" validate:"required,oneof=receive correction damage shrinkage cycle_count other"`
	Rows   []BulkInventoryAdjustRow `json:"rows" validate:"required,min=1,max=500"`
}

type BulkInventoryAdjustRowResult struct {
	SKU       string `json:"sku"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	ProductID *uint  `json:"product_id,omitempty"`
	Stock     *int   `json:"stock,omitempty"`
}

type BulkAdjustInventoryResponse struct {
	Applied int                            `json:"applied"`
	Failed  int                            `json:"failed"`
	Rows    []BulkInventoryAdjustRowResult `json:"rows"`
}

type ListInventoryHistoryRequest struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=20"`
}

type InventoryOverviewResponse struct {
	LowStockCount     int64 `json:"low_stock_count"`
	OutOfStockCount   int64 `json:"out_of_stock_count"`
	NotTrackedCount   int64 `json:"not_tracked_count"`
	TrackedSKUCount   int64 `json:"tracked_sku_count"`
	TotalUnitsOnHand  int64 `json:"total_units_on_hand"`
	WaitlistTotal     int64 `json:"waitlist_total"`
}

type InventoryItemResponse struct {
	ID                uint       `json:"id"`
	Name              string     `json:"name"`
	SKU               string     `json:"sku"`
	Slug              string     `json:"slug"`
	ImageURL          string     `json:"image_url,omitempty"`
	Stock             int        `json:"stock"`
	LowStockThreshold int        `json:"low_stock_threshold"`
	TrackInventory    bool       `json:"track_inventory"`
	AllowBackorder    bool       `json:"allow_backorder"`
	WarehouseLocation string     `json:"warehouse_location,omitempty"`
	Status            string     `json:"status"`
	WorkflowState     *StateView `json:"workflow_state,omitempty"`
	WaitlistCount     int64      `json:"waitlist_count"`
	UnitsSold30d      int64      `json:"units_sold_30d"`
	StockStatus       string     `json:"stock_status"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type InventoryListData struct {
	Items []InventoryItemResponse `json:"items"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
}

type InventoryAdjustmentResponse struct {
	ID             uint      `json:"id"`
	ProductID      uint      `json:"product_id"`
	ProductName    string    `json:"product_name,omitempty"`
	ProductSKU     string    `json:"product_sku,omitempty"`
	QuantityDelta  int       `json:"quantity_delta"`
	QuantityBefore int       `json:"quantity_before"`
	QuantityAfter  int       `json:"quantity_after"`
	AdjustmentType string    `json:"adjustment_type"`
	ReferenceType  string    `json:"reference_type,omitempty"`
	ReferenceID    *uint     `json:"reference_id,omitempty"`
	ActorName      string    `json:"actor_name,omitempty"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
}

type InventoryHistoryData struct {
	Adjustments []InventoryAdjustmentResponse `json:"adjustments"`
	Total       int64                         `json:"total"`
	Page        int                           `json:"page"`
	Limit       int                           `json:"limit"`
}
