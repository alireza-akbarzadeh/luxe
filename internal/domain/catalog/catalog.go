package catalog

import "errors"

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductName = errors.New("product name is required")
	ErrInvalidPrice       = errors.New("product price must be positive")
	ErrInvalidSKU         = errors.New("product sku is required")
)

// Product is the catalog aggregate root (domain view).
type Product struct {
	ID          uint
	Name        string
	Slug        string
	Description string
	PriceCents  int64
	Currency    string
	SKU         string
	Stock       int
	Status      string
	CategoryID  *uint
	BrandID     *uint
}

// CreateProductInput is domain input for creating a product.
type CreateProductInput struct {
	Name        string
	Description string
	PriceCents  int64
	Currency    string
	SKU         string
	Stock       int
	CategoryID  *uint
	BrandID     *uint
}

// Service holds catalog domain rules.
type Service struct{}

// NewService creates a catalog domain service.
func NewService() *Service { return &Service{} }

// ValidateCreate enforces invariants before persistence.
func (s *Service) ValidateCreate(in CreateProductInput) error {
	if in.Name == "" {
		return ErrInvalidProductName
	}
	if in.PriceCents <= 0 {
		return ErrInvalidPrice
	}
	if in.SKU == "" {
		return ErrInvalidSKU
	}
	return nil
}

// CanPublish reports whether a product may be published.
func (s *Service) CanPublish(p Product) error {
	if p.PriceCents <= 0 {
		return ErrInvalidPrice
	}
	if p.Name == "" {
		return ErrInvalidProductName
	}
	return nil
}
