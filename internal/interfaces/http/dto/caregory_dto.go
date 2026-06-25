package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type OrderFilters struct {
	Status    string
	FromDate  *time.Time
	ToDate    *time.Time
	MinAmount *float64
	MaxAmount *float64
}

type OrderListFilters struct {
	Limit     int        `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int        `form:"offset" validate:"omitempty,min=0"`
	Status    string     `form:"status" validate:"omitempty,oneof=pending paid shipped delivered cancelled refunded"`
	FromDate  *time.Time `form:"from_date"`
	ToDate    *time.Time `form:"to_date"`
	MinAmount *float64   `form:"min_amount" validate:"omitempty,gt=0"`
	MaxAmount *float64   `form:"max_amount" validate:"omitempty,gt=0"`
}

type CategorySingleResponse struct {
	BaseResponse
	Data CategoryData `json:"data"`
}

type CategoryData struct {
	Category models.Category `json:"category"`
}

type CategoryListResponse struct {
	BaseResponse
	Data CategoryListData `json:"data"`
}

type CategoryListData struct {
	Categories []models.Category `json:"categories"`
	Total      int64             `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
}

// BulkCreateCategoryResponse – for bulk create
type BulkCreateCategoryResponse struct {
	BaseResponse
	Data BulkCategoryData `json:"data"`
}

type BulkCategoryData struct {
	Categories []*models.Category `json:"categories"`
}
