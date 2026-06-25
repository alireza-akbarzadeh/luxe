package catalog

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
