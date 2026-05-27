package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type StoreFilter struct {
	Search       string  `form:"search"`
	Location     string  `form:"location"`
	MinRating    float64 `form:"min_rating"`
	CategorySlug string  `form:"category_slug"`
	SortBy       string  `form:"sort_by"`
	Limit        int     `form:"limit"`
	Offset       int     `form:"offset"`
}

type CreateStoreRequest struct {
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description"`
	LogoURL      string `json:"logo_url"`
	BannerURL    string `json:"banner_url"`
	Location     string `json:"location"`
	ShippingInfo string `json:"shipping_info"`
	ReturnPolicy string `json:"return_policy"`
	UserID       *uint  `json:"user_id,omitempty"`
	CategoryIDs  []uint `json:"category_ids"`
}

type UpdateStoreRequest struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	LogoURL      *string `json:"logo_url"`
	BannerURL    *string `json:"banner_url"`
	Location     *string `json:"location"`
	ShippingInfo *string `json:"shipping_info"`
	ReturnPolicy *string `json:"return_policy"`
	Status       *string `json:"status"`
	IsVerified   *bool   `json:"is_verified"`
	CategoryIDs  *[]uint `json:"category_ids"`
}

type StoreResponse struct {
	ID            uint               `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	Description   string             `json:"description"`
	LogoURL       string             `json:"logo_url"`
	BannerURL     string             `json:"banner_url"`
	Location      string             `json:"location"`
	ShippingInfo  string             `json:"shipping_info"`
	ReturnPolicy  string             `json:"return_policy"`
	Rating        float64            `json:"rating"`
	ReviewCount   int                `json:"review_count"`
	FollowerCount int                `json:"follower_count"`
	IsVerified    bool               `json:"is_verified"`
	JoinedAt      time.Time          `json:"joined_at"`
	Categories    []CategoryResponse `json:"categories,omitempty"`
}

func ToStoreResponse(store *models.Store) StoreResponse {
	cats := make([]CategoryResponse, len(store.Categories))
	for i, c := range store.Categories {
		cats[i] = CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		}
	}
	return StoreResponse{
		ID:            store.ID,
		Name:          store.Name,
		Slug:          store.Slug,
		Description:   store.Description,
		LogoURL:       store.LogoURL,
		BannerURL:     store.BannerURL,
		Location:      store.Location,
		ShippingInfo:  store.ShippingInfo,
		ReturnPolicy:  store.ReturnPolicy,
		Rating:        store.Rating,
		ReviewCount:   store.ReviewCount,
		FollowerCount: store.FollowerCount,
		IsVerified:    store.IsVerified,
		JoinedAt:      store.JoinedAt,
		Categories:    cats,
	}
}
