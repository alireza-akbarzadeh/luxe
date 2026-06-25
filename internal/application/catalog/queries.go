package catalog

import (
	"context"

	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/catalog"
	domainshared "github.com/alireza-akbarzadeh/luxe/internal/domain/shared"
)

// Queries orchestrates catalog read use cases.
type Queries struct {
	repo   domain.ProductRepository
	reader ProductReader
}

// NewQueries creates catalog query use cases.
func NewQueries(repo domain.ProductRepository, reader ProductReader) *Queries {
	return &Queries{repo: repo, reader: reader}
}

// GetByID loads a product by primary key.
func (q *Queries) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	return q.repo.GetByID(ctx, id)
}

// GetBySlug loads a product by slug.
func (q *Queries) GetBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return q.repo.GetBySlug(ctx, slug)
}

// List returns a paginated product list.
func (q *Queries) List(ctx context.Context, filter domain.ListFilter, page domainshared.PageParams) ([]domain.Product, int64, error) {
	normalized := page.Normalize(20, 100)
	return q.repo.List(ctx, filter, normalized.Limit, normalized.Offset)
}

// Commands orchestrates catalog write validation and persistence.
type Commands struct {
	domain *domain.Service
	repo   domain.ProductRepository
	writer ProductWriter
}

// NewCommands creates catalog command use cases.
func NewCommands(domainSvc *domain.Service, repo domain.ProductRepository, writer ProductWriter) *Commands {
	return &Commands{domain: domainSvc, repo: repo, writer: writer}
}

// ValidateCreate runs domain rules before legacy service persists.
func (c *Commands) ValidateCreate(in domain.CreateProductInput) error {
	return c.domain.ValidateCreate(in)
}

// SlugTaken checks slug uniqueness.
func (c *Commands) SlugTaken(ctx context.Context, slug string, excludeID uint) (bool, error) {
	return c.repo.SlugExists(ctx, slug, excludeID)
}
