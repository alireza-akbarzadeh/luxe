// Package errors provides domain-oriented error kinds mapped to HTTP AppError.
package errors

import (
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Domain sentinel errors — use errors.Is in domain/application layers.
var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// ToAppError maps domain errors to HTTP-layer AppError for handlers.
func ToAppError(err error) *utils.AppError {
	if err == nil {
		return nil
	}
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	switch {
	case errors.Is(err, ErrNotFound):
		return utils.ErrNotFound(err.Error())
	case errors.Is(err, ErrInvalidInput):
		return utils.ErrValidationFailed(err.Error())
	case errors.Is(err, ErrConflict):
		return utils.ErrConflict(err.Error())
	case errors.Is(err, ErrUnauthorized):
		return utils.ErrUnauthorized(err.Error())
	case errors.Is(err, ErrForbidden):
		return utils.ErrForbidden(err.Error())
	default:
		return utils.ErrInternal(err)
	}
}
