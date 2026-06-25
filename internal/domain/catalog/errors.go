package catalog

import "errors"

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductName = errors.New("product name is required")
	ErrInvalidPrice       = errors.New("product price must be positive")
	ErrInvalidSKU         = errors.New("product sku is required")
)
