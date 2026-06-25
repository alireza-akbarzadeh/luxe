package collection

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Queries orchestrates collection read use cases.
type Queries struct {
	repo *postgres.CollectionRepository
}

// NewQueries creates collection query use cases.
func NewQueries(repo *postgres.CollectionRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a collection with workflow state.
func (q *Queries) GetByID(ctx context.Context, id uint) (*models.Collection, error) {
	return q.repo.GetByID(ctx, id)
}

// List returns paginated collections with normalized paging.
func (q *Queries) List(ctx context.Context, req *dto.ListCollectionsRequest) ([]models.Collection, int64, int, int, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	items, total, err := q.repo.List(ctx, req, page, limit)
	return items, total, page, limit, err
}
