package cart

// AddItemRequest is the HTTP payload for adding a cart line item.
type AddItemRequest struct {
	ProductID uint   `json:"product_id" validate:"required,gt=0"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
	Color     string `json:"color"`
	Size      string `json:"size"`
}

// UpdateCartItemRequest is the HTTP payload for updating a cart line item.
type UpdateCartItemRequest struct {
	Quantity int    `json:"quantity" validate:"omitempty,gt=0"`
	Color    string `json:"color"`
	Size     string `json:"size"`
}
