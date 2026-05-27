package dto

import "gorm.io/datatypes"

type AddItemResponse struct {
	BaseResponse
	Data CartItemData `json:"data"`
}

type CartItemData struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type GetCartResponse struct {
	BaseResponse
	Data GetCartData `json:"data"`
}

type GetCartData struct {
	Cart  CartData `json:"cart"`
	Total float64  `json:"total"`
}

type CartData struct {
	ID    uint             `json:"id"`
	Items []CartItemDetail `json:"items"`
}

type CartItemDetail struct {
	ID            uint    `json:"id"`
	ProductID     uint    `json:"product_id"`
	Name          string  `json:"name"`
	Quantity      int     `json:"quantity"`
	Price         float64 `json:"price"`
	Total         float64 `json:"total"`
	SelectedColor string  `gorm:"size:50" json:"selected_color"`
	SelectedSize  string  `gorm:"size:50" json:"selected_size"`
	// New fields
	Image         string         `json:"image,omitempty"`
	OriginalPrice float64        `json:"original_price,omitempty"`
	Color         datatypes.JSON `json:"color,omitempty"`
	Size          datatypes.JSON `json:"size,omitempty"`
	Discount      float64        `json:"discount"`

	Stock       int    `json:"stock"`
	IsInStock   bool   `json:"is_in_stock"`
	ProductName string `json:"product_name"`
}
