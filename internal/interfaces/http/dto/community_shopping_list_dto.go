package dto

// CommunityShoppingListItemResponse is a product on a community shopping list.
type CommunityShoppingListItemResponse struct {
	ID      uint            `json:"id"`
	Note    string          `json:"note,omitempty"`
	Product HomeProductItem `json:"product"`
}

// CommunityShoppingListResponse is a full community list with items.
type CommunityShoppingListResponse struct {
	ID            uint                                `json:"id"`
	Slug          string                              `json:"slug"`
	Title         string                              `json:"title"`
	Description   string                              `json:"description,omitempty"`
	Theme         string                              `json:"theme,omitempty"`
	CoverImageURL string                              `json:"cover_image_url,omitempty"`
	AuthorName    string                              `json:"author_name,omitempty"`
	AuthorHandle  string                              `json:"author_handle,omitempty"`
	Items         []CommunityShoppingListItemResponse `json:"items"`
}

// CommunityShoppingListListItem is a summary card for discovery pages.
type CommunityShoppingListListItem struct {
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

// CommunityShoppingListListResponse wraps active community shopping lists.
type CommunityShoppingListListResponse struct {
	Lists []CommunityShoppingListListItem `json:"lists"`
}
