package dto

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type SearchRequest struct {
	Query        string  `form:"q"`
	Limit        int     `form:"limit,default=20"`
	Offset       int     `form:"offset,default=0"`
	CategoryID   *uint   `form:"category_id"`
	CategorySlug string  `form:"category_slug"`
	StoreID      *uint   `form:"store_id"`
	MinPrice     float64 `form:"min_price"`
	MaxPrice     float64 `form:"max_price"`
	MinRating    float64 `form:"min_rating"`
	IsDigital    *bool   `form:"is_digital"`
	IsNew        *bool   `form:"is_new"`
	InStock      *bool   `form:"in_stock"`
	OnSale       *bool   `form:"on_sale"`
	Sort         string  `form:"sort"`
}

type SearchResponse struct {
	Products   []ProductResponse  `json:"products"`
	Stores     []StoreResponse    `json:"stores"`
	Categories []CategoryResponse `json:"categories"`
	Total      int64              `json:"total"`
}

type SuggestionItem struct {
	Type  string   `json:"type"`
	ID    *uint    `json:"id,omitempty"`
	Name  string   `json:"name"`
	Slug  string   `json:"slug,omitempty"`
	Image *string  `json:"image,omitempty"`
	Price *float64 `json:"price,omitempty"`
}

type SuggestionsResponse struct {
	Suggestions []SuggestionItem `json:"suggestions"`
}

type TrendingSearch struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

type TrendingResponse struct {
	Trending []TrendingSearch `json:"trending"`
}

func ToStoreResponses(ctx context.Context, stores []*models.Store) []StoreResponse {
	res := make([]StoreResponse, len(stores))
	for i, s := range stores {
		res[i] = ToStoreResponse(ctx, s)
	}
	return res
}

// ToCategoryResponse maps a single models.Category to a locale-aware CategoryResponse.
func ToCategoryResponse(ctx context.Context, c *models.Category) CategoryResponse {
	return resolvedCategoryResponse(ctx, c)
}

// ToCategoryResponses maps a slice of models.Category to []CategoryResponse.
func ToCategoryResponses(ctx context.Context, categories []*models.Category) []CategoryResponse {
	res := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		res[i] = ToCategoryResponse(ctx, cat)
	}
	return res
}
