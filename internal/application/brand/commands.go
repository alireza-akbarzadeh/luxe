package brand

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/brand"
)

// Commands orchestrates brand write use cases.
type Commands struct {
	domain *domain.Service
	repo   *postgres.BrandRepository
}

// NewCommands creates brand command use cases.
func NewCommands(domainSvc *domain.Service, repo *postgres.BrandRepository) *Commands {
	return &Commands{domain: domainSvc, repo: repo}
}

// Create inserts a new brand.
func (c *Commands) Create(ctx context.Context, req *dto.CreateBrandRequest) (*models.Brand, error) {
	if err := c.domain.ValidateName(req.Name); err != nil {
		return nil, err
	}
	brand := BuildCreateModel(req)
	if err := c.repo.Create(ctx, brand); err != nil {
		return nil, err
	}
	return brand, nil
}

// Update persists brand field changes.
func (c *Commands) Update(ctx context.Context, brand *models.Brand) error {
	return c.repo.Save(ctx, brand)
}

// Delete removes a brand by id.
func (c *Commands) Delete(ctx context.Context, id uint) (int64, error) {
	return c.repo.DeleteByID(ctx, id)
}
