package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ─── Request DTOs ────────────────────────────────────────────────────────────

type BulkStockUpdate struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Stock     int  `json:"stock" validate:"gte=0"`
}

type CreateProductRequest struct {
	Name              string   `json:"name" validate:"required,min=3,max=255"`
	Description       string   `json:"description,omitempty"`
	Price             float64  `json:"price" validate:"required,gte=0"`
	CompareAtPrice    *float64 `json:"compare_at_price,omitempty" validate:"omitempty,gte=0"`
	Cost              *float64 `json:"cost,omitempty" validate:"omitempty,gte=0"`
	SKU               string   `json:"sku" validate:"required,min=3,max=50"`
	Barcode           string   `json:"barcode,omitempty"`
	Stock             int      `json:"stock" validate:"gte=0"`
	LowStockThreshold int      `json:"low_stock_threshold,omitempty"`
	Weight            *float64 `json:"weight,omitempty" validate:"omitempty,gte=0"`
	IsDigital         bool     `json:"is_digital"`
	CategoryID        *uint    `json:"category_id,omitempty"`
	Images            []string `json:"images,omitempty"`
	Status            string   `json:"status" validate:"oneof=draft active inactive archived"`
	MetaTitle         string   `json:"meta_title,omitempty"`
	MetaDescription   string   `json:"meta_description,omitempty"`
	IsNew             *bool    `json:"is_new,omitempty"`
	Colors            []string `json:"colors,omitempty"`
	Sizes             []string `json:"sizes,omitempty"`
	StoreID           *uint    `json:"store_id,omitempty"`
}

type UpdateProductRequest struct {
	Name              *string   `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Description       *string   `json:"description,omitempty"`
	Price             *float64  `json:"price,omitempty" validate:"omitempty,gte=0"`
	CompareAtPrice    *float64  `json:"compare_at_price,omitempty" validate:"omitempty,gte=0"`
	Cost              *float64  `json:"cost,omitempty" validate:"omitempty,gte=0"`
	SKU               *string   `json:"sku,omitempty" validate:"omitempty,min=3,max=50"`
	Barcode           *string   `json:"barcode,omitempty"`
	Stock             *int      `json:"stock,omitempty" validate:"omitempty,gte=0"`
	LowStockThreshold *int      `json:"low_stock_threshold,omitempty"`
	Weight            *float64  `json:"weight,omitempty" validate:"omitempty,gte=0"`
	IsDigital         *bool     `json:"is_digital,omitempty"`
	CategoryID        *uint     `json:"category_id,omitempty"`
	Images            *[]string `json:"images,omitempty"`
	Status            *string   `json:"status,omitempty" validate:"omitempty,oneof=draft active inactive archived"`
	MetaTitle         *string   `json:"meta_title,omitempty"`
	MetaDescription   *string   `json:"meta_description,omitempty"`
	IsNew             *bool     `json:"is_new,omitempty"`
	Colors            *[]string `json:"colors,omitempty"`
	Sizes             *[]string `json:"sizes,omitempty"`
	StoreID           *uint     `json:"store_id,omitempty"`
}

type BulkDeleteProductsRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"required,min=1"`
}

type ProductListFilters struct {
	Status     string  `form:"status"`
	Name       string  `form:"name"`
	SKU        string  `form:"sku"`
	CategoryID uint    `form:"category_id"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	MinRating  float64 `form:"min_rating"`
	MaxRating  float64 `form:"max_rating"`
	MinReviews int     `form:"min_reviews"`
	MaxReviews int     `form:"max_reviews"`
	IsDigital  *bool   `form:"is_digital"`
	IsNew      *bool   `form:"is_new"`
	Sort       string  `form:"sort"`
	StoreID    *uint   `json:"store_id,omitempty"`
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

// CategoryResponse is a flat, GORM-free category shape for API responses.
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Level       int    `json:"level"`
	Path        string `json:"path,omitempty"`
	IsActive    bool   `json:"is_active"`
	ParentID    *uint  `json:"parent_id,omitempty"`
}

// ProductResponse is a flat, GORM-free product shape for API responses.
// Swag will generate clean types from this — no GormDeletedAt, no recursive models.
type ProductResponse struct {
	ID                uint              `json:"id"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	SKU               string            `json:"sku"`
	Barcode           string            `json:"barcode,omitempty"`
	Description       string            `json:"description,omitempty"`
	Price             float64           `json:"price"`
	CompareAtPrice    *float64          `json:"compare_at_price,omitempty"`
	Cost              *float64          `json:"cost,omitempty"`
	Stock             int               `json:"stock"`
	LowStockThreshold int               `json:"low_stock_threshold,omitempty"`
	Weight            *float64          `json:"weight,omitempty"`
	IsDigital         bool              `json:"is_digital"`
	IsNew             bool              `json:"is_new"`
	Status            string            `json:"status"`
	Rating            float64           `json:"rating"`
	ReviewsCount      int               `json:"reviews_count"`
	Images            []string          `json:"images,omitempty"`
	MetaTitle         string            `json:"meta_title,omitempty"`
	MetaDescription   string            `json:"meta_description,omitempty"`
	CategoryID        *uint             `json:"category_id,omitempty"`
	Category          *CategoryResponse `json:"category,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Colors            []string          `json:"colors,omitempty"`
	Sizes             []string          `json:"sizes,omitempty"`
}

// ToProductResponse maps a models.Product to a ProductResponse.
// Call this in controllers instead of passing *models.Product directly.
func ToProductResponse(p models.Product) ProductResponse {
	r := ProductResponse{
		ID:                p.ID,
		Name:              p.Name,
		Slug:              p.Slug,
		SKU:               p.SKU,
		Barcode:           p.Barcode,
		Description:       p.Description,
		Price:             p.Price,
		CompareAtPrice:    p.CompareAtPrice,
		Cost:              p.Cost,
		Stock:             p.Stock,
		LowStockThreshold: p.LowStockThreshold,
		Weight:            p.Weight,
		IsDigital:         p.IsDigital,
		IsNew:             p.IsNew,
		Status:            p.Status,
		Rating:            p.Rating,
		ReviewsCount:      p.ReviewsCount,
		Images:            p.Images,
		MetaTitle:         p.MetaTitle,
		MetaDescription:   p.MetaDescription,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		Colors:            p.Colors,
		Sizes:             p.Sizes,
	}

	if p.CategoryID != nil {
		r.CategoryID = p.CategoryID
	}

	if p.Category != nil && p.Category.ID != 0 {
		cat := CategoryResponse{
			ID:          p.Category.ID,
			Name:        p.Category.Name,
			Slug:        p.Category.Slug,
			Description: p.Category.Description,
			Level:       p.Category.Level,
			Path:        p.Category.Path,
			IsActive:    p.Category.IsActive,
			ParentID:    p.Category.ParentID,
		}
		r.Category = &cat
	}

	return r
}

// ToProductResponses maps a slice of *models.Product to []ProductResponse.
func ToProductResponses(products []*models.Product) []ProductResponse {
	result := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, ToProductResponse(*p))
	}
	return result

}

// ─── Envelope types (what Swag and Orval see for success responses) ──────────

// ProductSingleData wraps a single product response.
type ProductSingleData struct {
	Product ProductResponse `json:"product"`
}

type ProductWithLike struct {
	ProductResponse
	IsLiked bool `json:"is_liked"`
}

// ProductListData holds paginated product results.
// Using ProductResponse instead of models.Product to avoid GORM fields in Swagger.
type ProductListData struct {
	Products []ProductWithLike `json:"products"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// ProductSingleResponse defines the standard API envelope for a single product.
// @Description Standard success response envelope for a single product.
type ProductSingleResponse struct {
	// Data holds the actual response payload.
	Data ProductSingleData `json:"data"`
}

// ProductListResponse defines the standard API envelope for a product list.
// @Description Standard success response envelope for a product list.
type ProductListResponse struct {
	// Data holds the paginated product list.
	Data ProductListData `json:"data"`
}

type ProductListItem struct {
	Items   ProductResponse `json:"items"`
	IsLiked bool            `json:"is_liked"`
}

type SuggestionsRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"required,min=1,dive,gt=0"`
	Limit      int    `json:"limit" validate:"omitempty,min=1,max=20"`
}
