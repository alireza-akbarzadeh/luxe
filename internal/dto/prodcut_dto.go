package dto

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ─── Request DTOs ────────────────────────────────────────────────────────────

type BulkStockUpdate struct {
	ProductID uint `json:"product_id" validate:"required,gt=0"`
	Stock     int  `json:"stock" validate:"gte=0"`
}
type ProductAttributeInput struct {
	Name   string   `json:"name" validate:"required"`
	Values []string `json:"values" validate:"required,min=1"`
}

type CreateProductRequest struct {
	Name              string                  `json:"name" validate:"required,min=3,max=255"`
	NameI18n          i18n.LocalizedMap       `json:"nameI18n,omitempty"`
	Description       string                  `json:"description,omitempty"`
	DescriptionI18n   i18n.LocalizedMap       `json:"descriptionI18n,omitempty"`
	SearchAliases     []string                `json:"searchAliases,omitempty"`
	Price             float64                 `json:"price" validate:"required,gte=0"`
	CompareAtPrice    *float64                `json:"compare_at_price,omitempty" validate:"omitempty,gte=0"`
	Cost              *float64                `json:"cost,omitempty" validate:"omitempty,gte=0"`
	SKU               string                  `json:"sku" validate:"required,min=3,max=50"`
	Barcode           string                  `json:"barcode,omitempty"`
	Stock             int                     `json:"stock" validate:"gte=0"`
	LowStockThreshold int                     `json:"low_stock_threshold,omitempty"`
	Weight            *float64                `json:"weight,omitempty" validate:"omitempty,gte=0"`
	IsDigital         bool                    `json:"is_digital"`
	CategoryID        *uint                   `json:"category_id,omitempty"`
	Images            []string                `json:"images,omitempty"`
	Status            string                  `json:"status" validate:"oneof=draft active inactive archived"`
	MetaTitle         string                  `json:"meta_title,omitempty"`
	MetaDescription   string                  `json:"meta_description,omitempty"`
	IsNew             *bool                   `json:"is_new,omitempty"`
	Colors            []string                `json:"colors,omitempty"`
	Sizes             []string                `json:"sizes,omitempty"`
	StoreID           *uint                   `json:"store_id,omitempty"`
	BrandID           *uint                   `json:"brand_id,omitempty"`
	Attributes        []ProductAttributeInput `json:"attributes,omitempty"`
	TrackInventory    *bool                   `json:"track_inventory,omitempty"`
	WarehouseLocation string                  `json:"warehouse_location,omitempty"`
	AllowBackorder    *bool                   `json:"allow_backorder,omitempty"`

	Visibility  string     `json:"visibility,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	Channels    []string   `json:"channels,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type UpdateProductRequest struct {
	Name              *string            `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	NameI18n          i18n.LocalizedMap  `json:"nameI18n,omitempty"`
	Description       *string            `json:"description,omitempty"`
	DescriptionI18n   i18n.LocalizedMap  `json:"descriptionI18n,omitempty"`
	SearchAliases     *[]string          `json:"searchAliases,omitempty"`
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

	BrandID    *uint                    `json:"brand_id,omitempty"`
	Attributes *[]ProductAttributeInput `json:"attributes,omitempty"`

	TrackInventory    *bool   `json:"track_inventory,omitempty"`
	WarehouseLocation *string `json:"warehouse_location,omitempty"`
	AllowBackorder    *bool   `json:"allow_backorder,omitempty"`

	Visibility  *string    `json:"visibility,omitempty"`
	Tags        *[]string  `json:"tags,omitempty"`
	Channels    *[]string  `json:"channels,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
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
	BrandID    *uint   `json:"brand_id,omitempty"`
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

// CategoryResponse is a flat, GORM-free category shape for API responses.
type CategoryResponse struct {
	ID              uint              `json:"id"`
	Name            string            `json:"name"`
	NameI18n        i18n.LocalizedMap `json:"nameI18n,omitempty"`
	Slug            string            `json:"slug"`
	Description     string            `json:"description,omitempty"`
	DescriptionI18n i18n.LocalizedMap `json:"descriptionI18n,omitempty"`
	Level           int               `json:"level"`
	Path            string            `json:"path,omitempty"`
	IsActive        bool              `json:"is_active"`
	ParentID        *uint             `json:"parent_id,omitempty"`
}

// ProductResponse is a flat, GORM-free product shape for API responses.
// Swag will generate clean types from this — no GormDeletedAt, no recursive models.
type ProductResponse struct {
	ID                uint              `json:"id"`
	Name              string            `json:"name"`
	NameI18n          i18n.LocalizedMap `json:"nameI18n,omitempty"`
	Slug              string            `json:"slug"`
	SKU               string            `json:"sku"`
	Barcode           string            `json:"barcode,omitempty"`
	Description       string            `json:"description,omitempty"`
	DescriptionI18n   i18n.LocalizedMap `json:"descriptionI18n,omitempty"`
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

	BrandID    *uint                      `json:"brand_id,omitempty"`
	Brand      *BrandResponse             `json:"brand,omitempty"`
	Attributes []ProductAttributeResponse `json:"attributes,omitempty"`

	StoreID           uint                   `json:"store_id,omitempty"`
	Store             *ProductStoreSummary   `json:"store,omitempty"`
	TrackInventory    bool                   `json:"track_inventory"`
	WarehouseLocation string `json:"warehouse_location,omitempty"`
	AllowBackorder    bool   `json:"allow_backorder"`

	// Publishing extras
	Visibility  string     `json:"visibility"`
	Tags        []string   `json:"tags,omitempty"`
	Channels    []string   `json:"channels,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	WorkflowState *StateView `json:"workflow_state,omitempty"`
}

// ToProductResponse maps a models.Product to a locale-aware ProductResponse.
func ToProductResponse(ctx context.Context, p models.Product) ProductResponse {
	r := ProductResponse{
		ID:                p.ID,
		Slug:              p.Slug,
		SKU:               p.SKU,
		Barcode:           p.Barcode,
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
		TrackInventory:    p.TrackInventory,
		WarehouseLocation: p.WarehouseLocation,
		AllowBackorder:    p.AllowBackorder,
		Visibility:        p.Visibility,
		Tags:              p.Tags,
		Channels:          p.Channels,
		PublishedAt:       p.PublishedAt,
	}
	applyProductI18nFields(ctx, &p, &r)

	if p.CategoryID != nil {
		r.CategoryID = p.CategoryID
	}

	if p.BrandID != nil {
		r.BrandID = p.BrandID
	}
	if p.Brand != nil && p.Brand.ID != 0 {
		r.Brand = &BrandResponse{
			ID:          p.Brand.ID,
			Name:        p.Brand.Name,
			Slug:        p.Brand.Slug,
			LogoURL:     p.Brand.LogoURL,
			Description: p.Brand.Description,
			Status:      p.Brand.Status,
			CreatedAt:   p.Brand.CreatedAt,
			UpdatedAt:   p.Brand.UpdatedAt,
		}
	}
	if len(p.Attributes) > 0 {
		attrs := make([]ProductAttributeResponse, 0, len(p.Attributes))
		for _, a := range p.Attributes {
			attrs = append(attrs, ProductAttributeResponse{Name: a.Name, Values: a.Values})
		}
		r.Attributes = attrs
	}

	if p.Category != nil && p.Category.ID != 0 {
		cat := resolvedCategoryResponse(ctx, p.Category)
		r.Category = &cat
	}

	if p.StoreID != 0 {
		r.StoreID = p.StoreID
	}
	if p.Store != nil && p.Store.ID != 0 {
		r.Store = ToProductStoreSummary(p.Store)
	}

	if p.WorkflowState != nil {
		r.WorkflowState = ToStateView(p.WorkflowState)
	}

	return r
}

// ToProductResponses maps a slice of *models.Product to []ProductResponse.
func ToProductResponses(ctx context.Context, products []*models.Product) []ProductResponse {
	result := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		result = append(result, ToProductResponse(ctx, *p))
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

type ToggleLikeRequest struct {
	Like bool `json:"like"`
}

type ToggleLikeResponse struct {
	Liked bool `json:"liked"`
}

type ProductAttributeResponse struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// PerformProductTransitionRequest triggers a workflow event on a product (admin).
type PerformProductTransitionRequest struct {
	Event string `json:"event" validate:"required,min=1,max=64"`
	Note  string `json:"note"  validate:"omitempty,max=512"`
}

// ProductTransitionResponse is returned after a successful product workflow transition.
type ProductTransitionResponse struct {
	Transition TransitionResultView `json:"transition"`
	Product    ProductResponse      `json:"product"`
}
