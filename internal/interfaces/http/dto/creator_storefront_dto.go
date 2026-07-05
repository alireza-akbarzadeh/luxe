package dto

// CreatorPickResponse is a curated product on a creator storefront.
type CreatorPickResponse struct {
	ID       uint            `json:"id"`
	Headline string          `json:"headline,omitempty"`
	Product  HomeProductItem `json:"product"`
}

// CreatorStorefrontResponse is a full creator profile with curated picks.
type CreatorStorefrontResponse struct {
	ID            uint                  `json:"id"`
	Slug          string                `json:"slug"`
	DisplayName   string                `json:"display_name"`
	Handle        string                `json:"handle,omitempty"`
	Bio           string                `json:"bio,omitempty"`
	Specialty     string                `json:"specialty,omitempty"`
	AvatarURL     string                `json:"avatar_url,omitempty"`
	CoverImageURL string                `json:"cover_image_url,omitempty"`
	InstagramURL  string                `json:"instagram_url,omitempty"`
	Picks         []CreatorPickResponse `json:"picks"`
}

// CreatorStorefrontListItem is a summary card for discovery pages.
type CreatorStorefrontListItem struct {
	ID            uint   `json:"id"`
	Slug          string `json:"slug"`
	DisplayName   string `json:"display_name"`
	Handle        string `json:"handle,omitempty"`
	Bio           string `json:"bio,omitempty"`
	Specialty     string `json:"specialty,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	CoverImageURL string `json:"cover_image_url,omitempty"`
	PickCount     int    `json:"pick_count"`
}

// CreatorStorefrontListResponse wraps active creator storefronts.
type CreatorStorefrontListResponse struct {
	Creators []CreatorStorefrontListItem `json:"creators"`
}
