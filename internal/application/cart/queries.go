package cart

import (
	"context"
	"errors"

	domaincart "github.com/alireza-akbarzadeh/luxe/internal/domain/cart"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates cart read use cases.
type Queries struct {
	reader Reader
}

// NewQueries creates cart query use cases.
func NewQueries(reader Reader) *Queries {
	return &Queries{reader: reader}
}

// GetOrCreate returns the user's active cart, creating one when missing.
func (q *Queries) GetOrCreate(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := q.reader.FindActiveCart(ctx, userID, true)
	if err == nil {
		return cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}
	return nil, domaincart.ErrCartNotFound
}

// Get returns the user's active cart or an empty cart when none exists.
func (q *Queries) Get(ctx context.Context, userID uint) (*models.Cart, error) {
	cart, err := q.reader.FindActiveCart(ctx, userID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return cart, nil
}
