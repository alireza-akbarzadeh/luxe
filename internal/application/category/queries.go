package category

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Queries orchestrates category read use cases.
type Queries struct {
	repo *postgres.CategoryRepository
}

// NewQueries creates category query use cases.
func NewQueries(repo *postgres.CategoryRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a category with relations.
func (q *Queries) GetByID(ctx context.Context, id uint) (*models.Category, error) {
	return q.repo.GetByID(ctx, id)
}

// GetBySlug loads a category by slug with relations.
func (q *Queries) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	return q.repo.GetBySlug(ctx, slug)
}

// List returns filtered categories with pagination defaults applied.
func (q *Queries) List(ctx context.Context, filters dto.CategoryListFilters) ([]models.Category, int64, error) {
	if filters.Limit == 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	return q.repo.List(ctx, filters)
}
