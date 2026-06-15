package dto

type CompareRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"required,min=2,max=4,dive,gt=0"`
}

type CompareProductResponse struct {
	ProductResponse
	StoreName       string  `json:"store_name"`
	StoreSlug       string  `json:"store_slug"`
	StoreLogo       string  `json:"store_logo,omitempty"`
	ShippingInfo    string  `json:"shipping_info"`
	ReturnPolicy    string  `json:"return_policy"`
	DiscountPercent float64 `json:"discount_percent,omitempty"`
}

type SyncCompareRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"max=4,dive,gt=0"`
}

type CompareListResponse struct {
	ProductIDs []uint `json:"product_ids"`
}
