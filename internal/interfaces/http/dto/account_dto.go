package dto

import (
	"time"
)

type DefaultAddressDTO struct {
	ID           uint   `json:"id"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
}

type DashboardSummaryResponse struct {
	ID                     uint               `json:"id"`
	Email                  string             `json:"email"`
	FirstName              string             `json:"first_name"`
	LastName               string             `json:"last_name"`
	Phone                  string             `json:"phone"`
	AvatarURL              string             `json:"avatar_url,omitempty"`
	Role                   string             `json:"role"`
	IsActive               bool               `json:"is_active"`
	EmailVerifiedAt        *time.Time         `json:"email_verified_at,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	DefaultShippingAddress *DefaultAddressDTO `json:"default_shipping_address,omitempty"`
	DefaultBillingAddress  *DefaultAddressDTO `json:"default_billing_address,omitempty"`
	AddressCount           int                `json:"address_count"`
	LikedProductsCount     int                `json:"liked_products_count"`
	RecentOrders           []OrderResponse    `json:"recent_orders"`
	MembershipTier         string             `json:"membership_tier"`
	IsPlusActive           bool               `json:"is_plus_active"`
	PlusSubscribedAt       *time.Time         `json:"plus_subscribed_at,omitempty"`
	PlusExpiresAt          *time.Time         `json:"plus_expires_at,omitempty"`
}

type OrderItemDetailDTO struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

type OrderDetailDTO struct {
	ID          uint                 `json:"id"`
	OrderNumber string               `json:"order_number"`
	CreatedAt   time.Time            `json:"created_at"`
	Status      string               `json:"status"`
	TotalAmount float64              `json:"total_amount"`
	Items       []OrderItemDetailDTO `json:"items"`
}

type OrderListResponseData struct {
	Orders []OrderDetailDTO `json:"orders"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type WishlistItemDTO struct {
	ProductID       uint     `json:"product_id"`
	ProductName     string   `json:"product_name"`
	Price           float64  `json:"price"`
	OldPrice        *float64 `json:"old_price,omitempty"`
	DiscountPercent *int     `json:"discount_percent,omitempty"`
	IsInStock       bool     `json:"is_in_stock"`
	StockQuantity   int      `json:"stock_quantity"`
	IsActive        bool     `json:"is_active"`
	ImageURL        string   `json:"image_url"`
	Stock           int      `json:"stock"`
	Color           []string `json:"color,omitempty"`
	Size            []string `json:"size,omitempty"`
}

type WishlistResponseData struct {
	Items  []WishlistItemDTO `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
