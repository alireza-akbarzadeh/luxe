package wallet

import "context"

// Repository loads and persists wallet balances and transactions.
type Repository interface {
	GetByUserID(ctx context.Context, userID uint) (*Wallet, error)
}
