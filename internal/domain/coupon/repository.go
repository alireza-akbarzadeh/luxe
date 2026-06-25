package coupon

import "context"

// Repository loads and persists coupons.
type Repository interface {
	GetByCode(ctx context.Context, code string) (*Coupon, error)
}
