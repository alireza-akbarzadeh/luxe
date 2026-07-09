package brand

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Queries orchestrates brand read use cases.
type Queries struct {
	repo *postgres.BrandRepository
}

// NewQueries creates brand query use cases.
func NewQueries(repo *postgres.BrandRepository) *Queries {
	return &Queries{repo: repo}
}

// GetByID loads a brand with workflow state.
func (q *Queries) GetByID(ctx context.Context, id uint) (*models.Brand, error) {
	return q.repo.GetByID(ctx, id)
}

// List returns paginated brands.
func (q *Queries) List(ctx context.Context, req *dto.ListBrandsRequest) ([]postgres.BrandListItem, int64, error) {
	return q.repo.List(ctx, req)
}
