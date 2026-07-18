package returnorder

import (
	"errors"
	"time"
)

var (
	ErrOrderNotEligible    = errors.New("returns are only allowed for delivered or completed orders")
	ErrOpenReturnExists    = errors.New("an open return already exists for this order")
	ErrReturnWindowExpired = errors.New("the return window for this order has expired")
	ErrInvalidReturnType   = errors.New("return type must be refund or exchange")
)

const (
	StatusDelivered    = "delivered"
	StatusCompleted    = "completed"
	ReturnTypeRefund   = "refund"
	ReturnTypeExchange = "exchange"
)

// IsOrderEligibleForReturn reports whether an order status allows a return request.
func IsOrderEligibleForReturn(orderStatus string) bool {
	return orderStatus == StatusDelivered || orderStatus == StatusCompleted
}

// ValidateReturnType ensures the requested resolution type is supported.
func ValidateReturnType(returnType string) error {
	if returnType == "" {
		return nil
	}
	if returnType == ReturnTypeRefund || returnType == ReturnTypeExchange {
		return nil
	}
	return ErrInvalidReturnType
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

// ValidateReturnWindow ensures the order is still within the tier return window.
func ValidateReturnWindow(reference time.Time, windowDays int) error {
	if windowDays <= 0 {
		return nil
	}
	deadline := reference.Add(time.Duration(windowDays) * 24 * time.Hour)
	if time.Now().After(deadline) {
		return ErrReturnWindowExpired
	}
	return nil
}
