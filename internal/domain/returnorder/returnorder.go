package returnorder

import "errors"

var (
	ErrOrderNotEligible = errors.New("returns are only allowed for delivered or completed orders")
	ErrOpenReturnExists = errors.New("an open return already exists for this order")
)

const (
	StatusDelivered = "delivered"
	StatusCompleted = "completed"
)

// IsOrderEligibleForReturn reports whether an order status allows a return request.
func IsOrderEligibleForReturn(orderStatus string) bool {
	return orderStatus == StatusDelivered || orderStatus == StatusCompleted
}

// ValidateCreateRequest checks whether a new return may be created.
func ValidateCreateRequest(orderStatus string, openReturnCount int64) error {
	if !IsOrderEligibleForReturn(orderStatus) {
		return ErrOrderNotEligible
	}
	if openReturnCount > 0 {
		return ErrOpenReturnExists
	}
	return nil
}
