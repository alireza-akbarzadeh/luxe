package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

type CampaignPlacementsRequest struct {
	FlashDealIDs  []uint `json:"flash_deal_ids,omitempty"`
	SectionIDs    []uint `json:"section_ids,omitempty"`
	CollectionIDs []uint `json:"collection_ids,omitempty"`
}

// --- Flash deals ---

type AdminFlashDealListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all active draft ended"`
	Search string `form:"search" validate:"omitempty"`
}

type CreateFlashDealRequest struct {
	ProductID     uint       `json:"product_id" validate:"required,gt=0"`
	Title         string     `json:"title,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        time.Time  `json:"ends_at" validate:"required"`
	QuantityLimit *int       `json:"quantity_limit,omitempty" validate:"omitempty,gte=1"`
	SortOrder     int        `json:"sort_order"`
	Status        string     `json:"status" validate:"omitempty,oneof=active draft ended"`
}

type UpdateFlashDealRequest struct {
	ProductID     *uint      `json:"product_id,omitempty" validate:"omitempty,gt=0"`
	Title         *string    `json:"title,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
	QuantityLimit *int       `json:"quantity_limit,omitempty" validate:"omitempty,gte=1"`
	SortOrder     *int       `json:"sort_order,omitempty"`
	Status        *string    `json:"status,omitempty" validate:"omitempty,oneof=active draft ended"`
}

type FlashDealListResponse struct {
	BaseResponse
	Data FlashDealListData `json:"data"`
}

type FlashDealListData struct {
	Deals  []models.FlashDeal `json:"deals"`
	Total  int64              `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

type FlashDealSingleResponse struct {
	BaseResponse
	Data FlashDealData `json:"data"`
}

type FlashDealData struct {
	Deal models.FlashDeal `json:"deal"`
}

// --- Homepage sections ---

type AdminHomepageSectionListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all draft published archived"`
	Search string `form:"search" validate:"omitempty"`
}

type CreateHomepageSectionRequest struct {
	SectionKey string                 `json:"section_key" validate:"required,min=2,max=64"`
	Title      string                 `json:"title" validate:"required,min=2,max=255"`
	Href       string                 `json:"href" validate:"required,max=512"`
	ImageURL   string                 `json:"image_url"`
	SortOrder  int                    `json:"sort_order"`
	Status     string                 `json:"status" validate:"omitempty,oneof=draft published archived"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

type UpdateHomepageSectionRequest struct {
	SectionKey *string                `json:"section_key,omitempty" validate:"omitempty,min=2,max=64"`
	Title      *string                `json:"title,omitempty" validate:"omitempty,min=2,max=255"`
	Href       *string                `json:"href,omitempty" validate:"omitempty,max=512"`
	ImageURL   *string                `json:"image_url,omitempty"`
	SortOrder  *int                   `json:"sort_order,omitempty"`
	Status     *string                `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

type HomepageSectionListResponse struct {
	BaseResponse
	Data HomepageSectionListData `json:"data"`
}

type HomepageSectionListData struct {
	Sections []models.HomepageSection `json:"sections"`
	Total    int64                    `json:"total"`
	Limit    int                      `json:"limit"`
	Offset   int                      `json:"offset"`
}

type HomepageSectionSingleResponse struct {
	BaseResponse
	Data HomepageSectionData `json:"data"`
}

type HomepageSectionData struct {
	Section models.HomepageSection `json:"section"`
}

// --- Campaigns ---

type AdminCampaignListFilters struct {
	Limit  int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset int    `form:"offset" validate:"omitempty,min=0"`
	Status string `form:"status" validate:"omitempty,oneof=all draft scheduled active ended archived"`
	Search string `form:"search" validate:"omitempty"`
}

type CreateCampaignRequest struct {
	Name        string                     `json:"name" validate:"required,min=2,max=255"`
	Slug        string                     `json:"slug,omitempty" validate:"omitempty,min=2,max=128"`
	Description string                     `json:"description,omitempty"`
	StartsAt    *time.Time                 `json:"starts_at,omitempty"`
	EndsAt      *time.Time                 `json:"ends_at,omitempty"`
	Status      string                     `json:"status" validate:"omitempty,oneof=draft scheduled active ended archived"`
	Placements  *CampaignPlacementsRequest `json:"placements,omitempty"`
}

type UpdateCampaignRequest struct {
	Name        *string                    `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Slug        *string                    `json:"slug,omitempty" validate:"omitempty,min=2,max=128"`
	Description *string                    `json:"description,omitempty"`
	StartsAt    *time.Time                 `json:"starts_at,omitempty"`
	EndsAt      *time.Time                 `json:"ends_at,omitempty"`
	Status      *string                    `json:"status,omitempty" validate:"omitempty,oneof=draft scheduled active ended archived"`
	Placements  *CampaignPlacementsRequest `json:"placements,omitempty"`
}

type CampaignListResponse struct {
	BaseResponse
	Data CampaignListData `json:"data"`
}

type CampaignListData struct {
	Campaigns []models.Campaign `json:"campaigns"`
	Total     int64             `json:"total"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
}

type CampaignSingleResponse struct {
	BaseResponse
	Data CampaignData `json:"data"`
}

type CampaignData struct {
	Campaign models.Campaign `json:"campaign"`
}

type PromotionsKPIResponse struct {
	BaseResponse
	Data PromotionsKPIData `json:"data"`
}

type PromotionsKPIData struct {
	ActiveFlashDeals      int64 `json:"active_flash_deals"`
	PublishedBanners      int64 `json:"published_banners"`
	ActiveCampaigns       int64 `json:"active_campaigns"`
	ScheduledCampaigns    int64 `json:"scheduled_campaigns"`
}
