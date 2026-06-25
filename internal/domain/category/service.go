package category

// Service holds category domain rules.
type Service struct{}

// NewService creates a category domain service.
func NewService() *Service {
	return &Service{}
}

// ValidateName ensures a category has a display name.
func (s *Service) ValidateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	return nil
}
