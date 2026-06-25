package returnorder

import "errors"

var ErrReturnNotAllowed = errors.New("return not allowed")

type Return struct {
	ID      uint
	OrderID uint
	Status  string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) CanApprove(r Return) error {
	if r.Status != "requested" {
		return ErrReturnNotAllowed
	}
	return nil
}
