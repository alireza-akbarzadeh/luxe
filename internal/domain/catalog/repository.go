package catalog

import "context"

// ListFilter filters product lists at the domain/persistence boundary.
type ListFilter struct {
	Search     string
	CategoryID *uint
	BrandID    *uint
	Status     string
}

// ProductRepository persists catalog products.
type ProductRepository interface {
	GetByID(ctx context.Context, id uint) (*Product, error)
	GetBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, filter ListFilter, limit, offset int) ([]Product, int64, error)
	SlugExists(ctx context.Context, slug string, excludeID uint) (bool, error)
}
