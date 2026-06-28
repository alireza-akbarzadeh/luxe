package dto

import "time"

// HomeCategoriesResponse is the payload for GET /home/categories.
type HomeCategoriesResponse struct {
	Popular           []HomeCategoryItem `json:"popular"`
	Featured          []HomeCategoryItem `json:"featured"`
	ForYou            []HomeCategoryItem `json:"for_you"`
	Favorite          []HomeCategoryItem `json:"favorite,omitempty"`
	PersonalizedRails []HomeProductRail  `json:"personalized_rails,omitempty"`
}

// HomeCategoryItem is a category card for the homepage.
type HomeCategoryItem struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Description  string `json:"description,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	ProductCount int64  `json:"product_count"`
}

// HomeProductRail is a titled product carousel derived from user preferences.
type HomeProductRail struct {
	Key      string            `json:"key"`
	Title    string            `json:"title"`
	Products []HomeProductItem `json:"products"`
}

// HomeProductItem extends product card data for homepage sections.
type HomeProductItem struct {
	ProductResponse
	UnitsSold      int64      `json:"units_sold,omitempty"`
	WishlistCount  int64      `json:"wishlist_count,omitempty"`
	FlashEndsAt    *time.Time `json:"flash_ends_at,omitempty"`
	QuantityLimit  *int       `json:"quantity_limit,omitempty"`
	DiscountPercent *int      `json:"discount_percent,omitempty"`
}

// HomeBrandItem is a brand card for the homepage.
type HomeBrandItem struct {
	BrandResponse
	BannerURL    string  `json:"banner_url,omitempty"`
	ProductCount int64   `json:"product_count"`
	UnitsSold    int64   `json:"units_sold"`
	Revenue      float64 `json:"revenue"`
	MinPrice     float64 `json:"min_price"`
	Rating       float64 `json:"rating"`
}

// HomeFlashDealItem is a flash deal with product details.
type HomeFlashDealItem struct {
	ID            uint       `json:"id"`
	Product       HomeProductItem `json:"product"`
	EndsAt        time.Time  `json:"ends_at"`
	QuantityLimit *int       `json:"quantity_limit,omitempty"`
}

// HomeStoreItem is a featured store card.
type HomeStoreItem struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	LogoURL      string  `json:"logo_url,omitempty"`
	BannerURL    string  `json:"banner_url,omitempty"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
	ProductCount int64   `json:"product_count"`
	UnitsSold    int64   `json:"units_sold"`
	Revenue      float64 `json:"revenue"`
}

// HomeCollectionItem is a curated collection card.
type HomeCollectionItem struct {
	ID          uint   `json:"id"`
	Slug        string `json:"slug"`
	Eyebrow     string `json:"eyebrow,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Href        string `json:"href"`
	ImageURL    string `json:"image_url,omitempty"`
	CTALabel    string `json:"cta_label,omitempty"`
}

// HomeSectionItem is a seasonal or curated homepage block.
type HomeSectionItem struct {
	Key      string                 `json:"key"`
	Title    string                 `json:"title"`
	Href     string                 `json:"href"`
	ImageURL string                 `json:"image_url,omitempty"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// HomeShopByPriceItem is a price bucket card linking to filtered shop.
type HomeShopByPriceItem struct {
	Key       string  `json:"key"`
	Title     string  `json:"title"`
	MaxPrice  float64 `json:"max_price"`
	Href      string  `json:"href"`
	PreviewCount int64 `json:"preview_count,omitempty"`
}

// HomeManifestResponse lists available homepage section keys (lightweight).
type HomeManifestResponse struct {
	Sections []string `json:"sections"`
	CacheTTL int      `json:"cache_ttl_seconds"`
}

// SetFavoriteCategoriesRequest replaces the user's favorite categories.
type SetFavoriteCategoriesRequest struct {
	CategoryIDs []uint `json:"category_ids" binding:"required"`
}

// FavoriteCategoriesResponse lists saved favorite categories.
type FavoriteCategoriesResponse struct {
	Categories []HomeCategoryItem `json:"categories"`
}

// HomeProductsResponse wraps a product list for homepage endpoints.
type HomeProductsResponse struct {
	Products []HomeProductItem `json:"products"`
}

// HomeBrandsResponse wraps brand list for homepage.
type HomeBrandsResponse struct {
	Brands []HomeBrandItem `json:"brands"`
}

// HomeFlashDealsResponse wraps flash deals.
type HomeFlashDealsResponse struct {
	Deals []HomeFlashDealItem `json:"deals"`
}

// HomeStoresResponse wraps featured stores.
type HomeStoresResponse struct {
	Stores []HomeStoreItem `json:"stores"`
}

// HomeCollectionsResponse wraps popular collections.
type HomeCollectionsResponse struct {
	Collections []HomeCollectionItem `json:"collections"`
}

// HomeSectionsResponse wraps seasonal/config sections.
type HomeSectionsResponse struct {
	Sections []HomeSectionItem `json:"sections"`
}

// HomeShopByPriceResponse wraps price bucket cards.
type HomeShopByPriceResponse struct {
	Buckets []HomeShopByPriceItem `json:"buckets"`
}
