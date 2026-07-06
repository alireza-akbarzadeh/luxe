package dto

// PublicCollectionItemResponse is a product on a community public collection.
type PublicCollectionItemResponse struct {
	ID      uint            `json:"id"`
	Note    string          `json:"note,omitempty"`
	Product HomeProductItem `json:"product"`
}

// PublicCollectionResponse is a full public collection with items.
type PublicCollectionResponse struct {
	ID            uint                           `json:"id"`
	Slug          string                         `json:"slug"`
	Title         string                         `json:"title"`
	Description   string                         `json:"description,omitempty"`
	Theme         string                         `json:"theme,omitempty"`
	CoverImageURL string                         `json:"cover_image_url,omitempty"`
	AuthorName    string                         `json:"author_name,omitempty"`
	AuthorHandle  string                         `json:"author_handle,omitempty"`
	Items         []PublicCollectionItemResponse `json:"items"`
}

// PublicCollectionListItem is a summary card for discovery pages.
type PublicCollectionListItem struct {
	ID            uint   `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	Theme         string `json:"theme,omitempty"`
	CoverImageURL string `json:"cover_image_url,omitempty"`
	AuthorName    string `json:"author_name,omitempty"`
	AuthorHandle  string `json:"author_handle,omitempty"`
	ItemCount     int    `json:"item_count"`
}

// PublicCollectionListResponse wraps active public collections.
type PublicCollectionListResponse struct {
	Collections []PublicCollectionListItem `json:"collections"`
}
