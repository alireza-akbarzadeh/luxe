package collection

// Service holds collection domain rules.
type Service struct{}

// NewService creates a collection domain service.
func NewService() *Service {
	return &Service{}
}

// ValidateTitle ensures a collection has a title.
func (s *Service) ValidateTitle(title string) error {
	if title == "" {
		return ErrInvalidTitle
	}
	return nil
}
