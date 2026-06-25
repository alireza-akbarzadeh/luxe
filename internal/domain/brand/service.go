package brand

// Service holds brand domain rules.
type Service struct{}

// NewService creates a brand domain service.
func NewService() *Service {
	return &Service{}
}

// ValidateName ensures a brand has a display name.
func (s *Service) ValidateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	return nil
}
