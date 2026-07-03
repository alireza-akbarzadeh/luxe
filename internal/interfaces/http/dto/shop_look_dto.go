package dto

// ShopLookTagResponse is a hotspot with embedded product card data.
type ShopLookTagResponse struct {
	ID       uint            `json:"id"`
	XPercent float64         `json:"x_percent"`
	YPercent float64         `json:"y_percent"`
	Label    string          `json:"label,omitempty"`
	Product  HomeProductItem `json:"product"`
}

// ShopLookResponse is a full shop-the-look scene with tagged products.
type ShopLookResponse struct {
	ID          uint                  `json:"id"`
	Slug        string                `json:"slug"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	ImageURL    string                `json:"image_url"`
	Tags        []ShopLookTagResponse `json:"tags"`
}

// ShopLookListItem is a summary card for listing pages.
type ShopLookListItem struct {
	ID          uint   `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	TagCount    int    `json:"tag_count"`
}

// ShopLookListResponse wraps active shop looks.
type ShopLookListResponse struct {
	Looks []ShopLookListItem `json:"looks"`
}
