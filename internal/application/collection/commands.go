package collection

import (
	"context"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/collection"
)

// Commands orchestrates collection write use cases.
type Commands struct {
	domain *domain.Service
	repo   *postgres.CollectionRepository
}

// NewCommands creates collection command use cases.
func NewCommands(domainSvc *domain.Service, repo *postgres.CollectionRepository) *Commands {
	return &Commands{domain: domainSvc, repo: repo}
}

// UniqueSlug returns a slug that is not taken.
func (c *Commands) UniqueSlug(ctx context.Context, baseSlug string, excludeID uint) (string, error) {
	slug := baseSlug
	counter := 1
	for {
		taken, err := c.repo.SlugExists(ctx, slug, excludeID)
		if err != nil {
			return "", err
		}
		if !taken {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
}

// Create inserts a new collection and optional manual products.
func (c *Commands) Create(ctx context.Context, req *dto.CreateCollectionRequest, slug string) (*models.Collection, error) {
	if err := c.domain.ValidateTitle(req.Title); err != nil {
		return nil, err
	}
	collection, err := BuildCreateModel(req, slug)
	if err != nil {
		return nil, err
	}
	if err := c.repo.Create(ctx, collection); err != nil {
		return nil, err
	}
	if err := c.syncProducts(ctx, collection, collection.Mode, req.ProductIDs, req.ProductOverrides); err != nil {
		return nil, err
	}
	return collection, nil
}

// Update persists collection field changes and optional manual products.
func (c *Commands) Update(ctx context.Context, collection *models.Collection, req *dto.UpdateCollectionRequest, oldSlug string) error {
	if err := c.repo.Save(ctx, collection); err != nil {
		return err
	}
	if collection.Slug != oldSlug {
		if err := c.repo.UpsertSlugRedirect(ctx, collection.ID, oldSlug); err != nil {
			return err
		}
	}
	productIDs := productIDsForUpdate(collection, req)
	if productIDs != nil {
		overrides := ProductOverridesFromModel(collection)
		if req.ProductOverrides != nil {
			overrides = *req.ProductOverrides
		}
		return c.repo.ReplaceProducts(ctx, collection.ID, productIDs, overrides)
	}
	mode := collection.Mode
	if req.Mode != nil {
		mode = *req.Mode
	} else if req.CollectionType != nil && *req.CollectionType == "smart" {
		mode = "dynamic"
	}
	if mode == "dynamic" {
		return c.repo.ClearProducts(ctx, collection.ID)
	}
	return nil
}

// Delete removes a collection by id.
func (c *Commands) Delete(ctx context.Context, id uint) (int64, error) {
	return c.repo.DeleteByID(ctx, id)
}

// ActivateDue promotes scheduled collections that should go live.
func (c *Commands) ActivateDue(ctx context.Context) (int64, error) {
	return c.repo.ActivateDue(ctx)
}

// ExpireEnded deactivates collections past their ends_at window.
func (c *Commands) ExpireEnded(ctx context.Context) (int64, error) {
	return c.repo.ExpireEnded(ctx)
}

func productIDsForUpdate(collection *models.Collection, req *dto.UpdateCollectionRequest) []uint {
	if req.ProductIDs != nil {
		return *req.ProductIDs
	}
	if req.Mode != nil && (*req.Mode == "manual" || *req.Mode == "hybrid") {
		return ProductIDsFromModel(collection)
	}
	if req.CollectionType != nil && *req.CollectionType == "manual" {
		return ProductIDsFromModel(collection)
	}
	return nil
}

func (c *Commands) syncProducts(
	ctx context.Context,
	collection *models.Collection,
	mode string,
	productIDs []uint,
	overrides []dto.CollectionProductOverrideInput,
) error {
	if mode == "" {
		mode = collection.Mode
	}
	if mode == "" {
		mode = "dynamic"
	}
	if mode == "manual" || mode == "hybrid" {
		return c.repo.ReplaceProducts(ctx, collection.ID, productIDs, overrides)
	}
	return c.repo.ClearProducts(ctx, collection.ID)
}
