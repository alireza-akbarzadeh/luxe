package catalog

// Service holds catalog domain rules.
type Service struct{}

// NewService creates a catalog domain service.
func NewService() *Service {
	return &Service{}
}

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
