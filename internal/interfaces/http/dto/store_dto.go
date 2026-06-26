package dto

import (
	"context"
	"encoding/json"
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
	Name         string          `json:"name" validate:"required"`
	Description  string          `json:"description"`
	LogoURL      string          `json:"logo_url"`
	BannerURL    string          `json:"banner_url"`
	Location     string          `json:"location"`
	ShippingInfo string          `json:"shipping_info"`
	ReturnPolicy string          `json:"return_policy"`
	Status       string          `json:"status,omitempty" validate:"omitempty,oneof=active pending suspended"`
	UserID       *uint           `json:"user_id,omitempty"`
	CategoryIDs  []uint          `json:"category_ids"`
	Settings     json.RawMessage `json:"settings,omitempty"`
}

// VendorCreateStoreRequest is the self-service seller onboarding payload.
type VendorCreateStoreRequest struct {
	Name              string `json:"name" validate:"required"`
	Description       string `json:"description" validate:"required"`
	LogoURL           string `json:"logo_url"`
	BannerURL         string `json:"banner_url"`
	Location          string `json:"location" validate:"required"`
	ShippingInfo      string `json:"shipping_info" validate:"required"`
	ReturnPolicy      string `json:"return_policy" validate:"required"`
	BusinessLegalName string `json:"business_legal_name" validate:"required"`
	BusinessType      string `json:"business_type" validate:"required,oneof=individual company brand"`
	Country           string `json:"country" validate:"required"`
	Website           string `json:"website"`
	TaxID             string `json:"tax_id"`
	FulfillmentModel  string  `json:"fulfillment_model" validate:"required,oneof=self platform hybrid"`
	CategoryIDs       []uint  `json:"category_ids"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
}

// VendorUpdateStoreRequest updates a vendor-owned store profile.
type VendorUpdateStoreRequest struct {
	Name              *string  `json:"name"`
	Description       *string  `json:"description"`
	LogoURL           *string  `json:"logo_url"`
	BannerURL         *string  `json:"banner_url"`
	Location          *string  `json:"location"`
	ShippingInfo      *string  `json:"shipping_info"`
	ReturnPolicy      *string  `json:"return_policy"`
	BusinessLegalName *string  `json:"business_legal_name"`
	BusinessType      *string  `json:"business_type" validate:"omitempty,oneof=individual company brand"`
	Country           *string  `json:"country"`
	Website           *string  `json:"website"`
	TaxID             *string  `json:"tax_id"`
	FulfillmentModel  *string  `json:"fulfillment_model" validate:"omitempty,oneof=self platform hybrid"`
	CategoryIDs       *[]uint  `json:"category_ids"`
	Latitude          *float64 `json:"latitude,omitempty"`
	Longitude         *float64 `json:"longitude,omitempty"`
}

type AdminStoreFilter struct {
	Search     string `form:"search"`
	Status     string `form:"status"`
	IsVerified *bool  `form:"is_verified"`
	SortBy     string `form:"sort_by"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
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
	IsFollowed    *bool              `json:"is_followed,omitempty"`
}

// VendorStoreResponse includes vendor/admin fields not exposed on the public catalog.
type VendorStoreResponse struct {
	StoreResponse
	Status   string          `json:"status"`
	Settings json.RawMessage `json:"settings,omitempty"`
}

// AdminStoreResponse extends vendor view with ownership metadata.
type AdminStoreResponse struct {
	VendorStoreResponse
	UserID     *uint     `json:"user_id,omitempty"`
	OwnerEmail string    `json:"owner_email,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func ToVendorStoreResponse(ctx context.Context, store *models.Store) VendorStoreResponse {
	return VendorStoreResponse{
		StoreResponse: ToStoreResponse(ctx, store),
		Status:        store.Status,
		Settings:      json.RawMessage(store.Settings),
	}
}

func ToAdminStoreResponse(ctx context.Context, store *models.Store) AdminStoreResponse {
	resp := AdminStoreResponse{
		VendorStoreResponse: ToVendorStoreResponse(ctx, store),
		UserID:              store.UserID,
		CreatedAt:           store.CreatedAt,
	}
	if store.User != nil {
		resp.OwnerEmail = store.User.Email
	}
	return resp
}

func ToStoreResponse(ctx context.Context, store *models.Store) StoreResponse {
	cats := make([]CategoryResponse, len(store.Categories))
	for i := range store.Categories {
		cats[i] = ToCategoryResponse(ctx, store.Categories[i])
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
