package dto

type StorefrontCategoryItem struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	ProductCount int64  `json:"product_count"`
}

type StorefrontBrandItem struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	LogoURL      string `json:"logo_url,omitempty"`
	ProductCount int64  `json:"product_count"`
}

type StorefrontProductCard struct {
	ProductWithLike
	DiscountPercent *int  `json:"discount_percent,omitempty"`
	UnitsSold       int64 `json:"units_sold,omitempty"`
}
