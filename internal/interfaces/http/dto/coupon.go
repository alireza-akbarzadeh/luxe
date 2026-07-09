package dto

import (
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// CouponConditionsRequest holds optional eligibility rules for promotions.
type CouponConditionsRequest struct {
	FirstOrderOnly  bool   `json:"first_order_only,omitempty"`
	CategoryIDs     []uint `json:"category_ids,omitempty"`
	ProductIDs      []uint `json:"product_ids,omitempty"`
	UserIDs         []uint `json:"user_ids,omitempty"`
	CustomerSegment string `json:"customer_segment,omitempty" validate:"omitempty,oneof=vip plus new"`
	MinItemQuantity int    `json:"min_item_quantity,omitempty" validate:"omitempty,gte=1"`
}

type CreateCouponRequest struct {
	Code                   string                   `json:"code" validate:"omitempty,min=3,max=50"`
	Description            string                   `json:"description,omitempty"`
	ApplicationType        string                   `json:"application_type" validate:"omitempty,oneof=code automatic bogo"`
	DiscountType           string                   `json:"discount_type" validate:"required,oneof=percentage fixed"`
	DiscountValue          float64                  `json:"discount_value" validate:"required,gt=0"`
	MinimumOrderAmount     float64                  `json:"minimum_order_amount"`
	MaxDiscountAmount      *float64                 `json:"max_discount_amount,omitempty"`
	UsageLimit             int                      `json:"usage_limit"`
	IsActive               *bool                    `json:"is_active,omitempty"`
	StartDate              time.Time                `json:"start_date"`
	EndDate                time.Time                `json:"end_date"`
	Conditions             *CouponConditionsRequest `json:"conditions,omitempty"`
	BogoBuyQuantity        int                      `json:"bogo_buy_quantity" validate:"omitempty,gte=1"`
	BogoGetQuantity        int                      `json:"bogo_get_quantity" validate:"omitempty,gte=1"`
	BogoGetDiscountPercent float64                  `json:"bogo_get_discount_percent" validate:"omitempty,gte=0,lte=100"`
}

// AdminCouponListFilters supports admin coupon management listing.
type AdminCouponListFilters struct {
	Limit           int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset          int    `form:"offset" validate:"omitempty,min=0"`
	Code            string `form:"code" validate:"omitempty"`
	Status          string `form:"status" validate:"omitempty,oneof=active inactive expired exhausted all"`
	DiscountType    string `form:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	ApplicationType string `form:"application_type" validate:"omitempty,oneof=code automatic bogo"`
}

type UpdateCouponRequest struct {
	Code                   *string                  `json:"code,omitempty"`
	Description            *string                  `json:"description,omitempty"`
	ApplicationType        *string                  `json:"application_type,omitempty" validate:"omitempty,oneof=code automatic bogo"`
	DiscountType           *string                  `json:"discount_type,omitempty"`
	DiscountValue          *float64                 `json:"discount_value,omitempty"`
	MinimumOrderAmount     *float64                 `json:"minimum_order_amount,omitempty"`
	MaxDiscountAmount      *float64                 `json:"max_discount_amount,omitempty"`
	UsageLimit             *int                     `json:"usage_limit,omitempty"`
	StartDate              *time.Time               `json:"start_date,omitempty"`
	EndDate                *time.Time               `json:"end_date,omitempty"`
	IsActive               *bool                    `json:"is_active,omitempty"`
	Conditions             *CouponConditionsRequest `json:"conditions,omitempty"`
	BogoBuyQuantity        *int                     `json:"bogo_buy_quantity,omitempty" validate:"omitempty,gte=1"`
	BogoGetQuantity        *int                     `json:"bogo_get_quantity,omitempty" validate:"omitempty,gte=1"`
	BogoGetDiscountPercent *float64                 `json:"bogo_get_discount_percent,omitempty" validate:"omitempty,gte=0,lte=100"`
}

type CouponListFilters struct {
	Limit        int        `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset       int        `form:"offset" validate:"omitempty,min=0"`
	Code         string     `form:"code" validate:"omitempty"`
	IsActive     *bool      `form:"is_active"`
	DiscountType string     `form:"discount_type" validate:"omitempty,oneof=percentage fixed"`
	StartDate    *time.Time `form:"start_date"`
	EndDate      *time.Time `form:"end_date"`
}

type CouponSingleResponse struct {
	BaseResponse
	Data CouponData `json:"data"`
}

type CouponData struct {
	Coupon models.Coupon `json:"coupon"`
}

// CouponListResponse for paginated list
type CouponListResponse struct {
	BaseResponse
	Data CouponListData `json:"data"`
}

type CouponListData struct {
	Coupons []models.Coupon `json:"coupons"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
}

// CouponValidateResponse for validation endpoint
type CouponValidateResponse struct {
	BaseResponse
	Data CouponValidateData `json:"data"`
}

type CouponValidateData struct {
	Coupon         models.Coupon `json:"coupon"`
	DiscountAmount float64       `json:"discount_amount"`
	FinalTotal     float64       `json:"final_total"`
}

type ValidateRequest struct {
	Code       string  `json:"code" validate:"required"`
	OrderTotal float64 `json:"order_total" validate:"required,gt=0"`
	ItemCount  int     `json:"item_count" validate:"omitempty,gte=1"`
}

type BestAutomaticCouponQuery struct {
	OrderTotal float64 `form:"order_total" validate:"required,gt=0"`
	ItemCount  int     `form:"item_count" validate:"omitempty,gte=1"`
}
