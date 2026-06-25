package user

import "context"

// Repository loads and persists users.
type Repository interface {
	GetByID(ctx context.Context, id uint) (*User, error)
}
