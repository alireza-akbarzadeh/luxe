package shipment

import "errors"

var ErrInvalidShipmentState = errors.New("invalid shipment state")

type Shipment struct {
	ID      uint
	OrderID uint
	Status  string
}

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) CanMarkDelivered(sh Shipment) error {
	switch sh.Status {
	case "delivered", "cancelled", "returned":
		return ErrInvalidShipmentState
	default:
		return nil
	}
}
