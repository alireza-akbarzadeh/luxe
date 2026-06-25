package inventory

// Service holds inventory domain rules.
type Service struct{}

// NewService creates an inventory domain service.
func NewService() *Service {
	return &Service{}
}

// CanAdjust reports whether a delta is allowed for tracked inventory.
func (s *Service) CanAdjust(trackInventory bool, before, delta int, allowBackorder, skipCheck bool) error {
	if !trackInventory {
		return nil
	}
	after := before + delta
	if after < 0 {
		return ErrInsufficientStock
	}
	if delta < 0 && !skipCheck && !allowBackorder && after < 0 {
		return ErrInsufficientStock
	}
	return nil
}
