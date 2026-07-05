package returnorder

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates return read use cases.
type Queries struct {
	repo *postgres.ReturnRepository
}

// NewQueries creates return query use cases.
func NewQueries(repo *postgres.ReturnRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a return by id with optional admin scope.
func (q *Queries) GetByID(ctx context.Context, returnID, userID uint, isAdmin bool) (*models.Return, error) {
	ret, err := q.repo.FindByID(ctx, returnID, userID, isAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("return not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return ret, nil
}

// ListForUser returns paginated returns for a user.
func (q *Queries) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.Return, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	total, err := q.repo.CountForUser(ctx, userID)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	returns, err := q.repo.ListForUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return returns, total, nil
}

// ListAdmin returns paginated returns for admin.
func (q *Queries) ListAdmin(ctx context.Context, filters dto.AdminReturnListFilters) ([]models.Return, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	total, err := q.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	returns, err := q.repo.ListAdmin(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return returns, total, nil
}

// ProductReturnStats loads aggregate return signals for AI return-risk insights.
func (q *Queries) ProductReturnStats(ctx context.Context, productID uint) (orderCount int64, returnCount int64, reasons []string, err error) {
	orderCount, returnCount, reasons, err = q.repo.ProductReturnStats(ctx, productID)
	if err != nil {
		return 0, 0, nil, utils.ErrInternal(err)
	}
	return orderCount, returnCount, reasons, nil
}
