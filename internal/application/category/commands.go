package category

import (
	"context"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	domain "github.com/alireza-akbarzadeh/luxe/internal/domain/category"
	"gorm.io/gorm"
)

// Commands orchestrates category write use cases.
type Commands struct {
	domain *domain.Service
	repo   *postgres.CategoryRepository
}

// NewCommands creates category command use cases.
func NewCommands(domainSvc *domain.Service, repo *postgres.CategoryRepository) *Commands {
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

// UpdateLevelAndPath recalculates hierarchy fields from the parent.
func (c *Commands) UpdateLevelAndPath(ctx context.Context, category *models.Category) error {
	if category.ParentID == nil || *category.ParentID == 0 {
		category.Level = 0
		category.Path = ""
		return nil
	}
	parent, err := c.repo.FindParent(ctx, *category.ParentID)
	if err != nil {
		return err
	}
	category.Level = parent.Level + 1
	if parent.Path == "" {
		category.Path = fmt.Sprintf("%d", parent.ID)
	} else {
		category.Path = fmt.Sprintf("%s.%d", parent.Path, parent.ID)
	}
	return nil
}

// Create inserts a new category.
func (c *Commands) Create(ctx context.Context, req dto.CreateCategoryRequest, slug string) (*models.Category, error) {
	if err := c.domain.ValidateName(req.Name); err != nil {
		return nil, err
	}

	category := BuildCreateModel(req, slug)
	if err := c.UpdateLevelAndPath(ctx, category); err != nil {
		return nil, err
	}
	category.SearchDocument = dto.BuildCategorySearchDocument(category)

	if err := c.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// Save persists category updates and optionally refreshes product search docs.
func (c *Commands) Save(ctx context.Context, category *models.Category, categoryID uint) error {
	if err := c.repo.Save(ctx, category); err != nil {
		return err
	}
	c.refreshProductSearchDocuments(ctx, categoryID)
	return nil
}

func (c *Commands) refreshProductSearchDocuments(ctx context.Context, categoryID uint) {
	category, err := c.repo.GetByID(ctx, categoryID)
	if err != nil {
		return
	}
	products, err := c.repo.FindProductsByCategoryID(ctx, categoryID)
	if err != nil {
		return
	}
	for i := range products {
		doc := dto.BuildProductSearchDocument(&products[i], category)
		_ = c.repo.UpdateProductSearchDocument(ctx, products[i].ID, doc)
	}
}

// Delete removes a category when it has no children.
func (c *Commands) Delete(ctx context.Context, id uint) (int64, error) {
	count, err := c.repo.CountChildren(ctx, id)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrHasChildren
	}
	return c.repo.DeleteByID(ctx, id)
}

// BulkCreate inserts multiple categories in one transaction.
func (c *Commands) BulkCreate(ctx context.Context, requests []dto.CreateCategoryRequest, slugFn func(context.Context, string, uint) (string, error)) ([]*models.Category, error) {
	created := make([]*models.Category, 0, len(requests))
	modelsToCreate := make([]*models.Category, 0, len(requests))

	for _, req := range requests {
		slug := req.Slug
		if slug == "" {
			slug = GenerateSlug(req.Name)
		}
		unique, err := slugFn(ctx, slug, 0)
		if err != nil {
			return nil, err
		}

		cat := BuildCreateModel(req, unique)
		if err := c.UpdateLevelAndPath(ctx, cat); err != nil {
			return nil, err
		}
		cat.SearchDocument = dto.BuildCategorySearchDocument(cat)
		modelsToCreate = append(modelsToCreate, cat)
	}

	if err := c.repo.BulkCreate(ctx, modelsToCreate); err != nil {
		return nil, err
	}
	created = append(created, modelsToCreate...)
	return created, nil
}

// BulkDelete removes categories that have no children.
func (c *Commands) BulkDelete(ctx context.Context, ids []uint) (int64, error) {
	count, err := c.repo.CountChildrenIn(ctx, ids)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrHasChildren
	}
	return c.repo.BulkDeleteByIDs(ctx, ids)
}

func (c *Commands) isInvalidParentMove(ctx context.Context, category *models.Category, parentID *uint) (bool, error) {
	if parentID == nil || *parentID == 0 {
		return false, nil
	}
	if *parentID == category.ID {
		return true, nil
	}
	descendantIDs, err := c.repo.FindDescendantIDs(ctx, category)
	if err != nil {
		return false, err
	}
	for _, id := range descendantIDs {
		if id == *parentID {
			return true, nil
		}
	}
	return false, nil
}

func (c *Commands) refreshSubtreePaths(ctx context.Context, parentID uint) error {
	children, err := c.repo.ListChildren(ctx, &parentID)
	if err != nil {
		return err
	}
	for i := range children {
		child := children[i]
		if err := c.UpdateLevelAndPath(ctx, &child); err != nil {
			return err
		}
		if err := c.repo.Save(ctx, &child); err != nil {
			return err
		}
		if err := c.refreshSubtreePaths(ctx, child.ID); err != nil {
			return err
		}
	}
	return nil
}

// Reorder updates sibling order and optional parent moves for categories.
func (c *Commands) Reorder(ctx context.Context, req dto.ReorderCategoriesRequest) error {
	ids := make([]uint, len(req.Items))
	for i, item := range req.Items {
		ids[i] = item.ID
	}

	loaded, err := c.repo.FindByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if len(loaded) != len(req.Items) {
		return gorm.ErrRecordNotFound
	}

	byID := make(map[uint]*models.Category, len(loaded))
	for i := range loaded {
		byID[loaded[i].ID] = &loaded[i]
	}

	updates := make([]*models.Category, 0, len(req.Items))
	parentChanged := make(map[uint]bool)

	for _, item := range req.Items {
		category := byID[item.ID]
		if category == nil {
			return gorm.ErrRecordNotFound
		}

		invalid, err := c.isInvalidParentMove(ctx, category, item.ParentID)
		if err != nil {
			return err
		}
		if invalid {
			return ErrInvalidParentMove
		}

		if item.ParentID != nil && *item.ParentID > 0 {
			if _, err := c.repo.FindParent(ctx, *item.ParentID); err != nil {
				return err
			}
		}

		if !sameParentID(category.ParentID, item.ParentID) {
			parentChanged[category.ID] = true
			category.ParentID = item.ParentID
		}
		category.SortOrder = item.SortOrder
		updates = append(updates, category)
	}

	for _, category := range updates {
		if parentChanged[category.ID] {
			if err := c.UpdateLevelAndPath(ctx, category); err != nil {
				return err
			}
		}
	}

	if err := c.repo.Reorder(ctx, updates); err != nil {
		return err
	}

	for id := range parentChanged {
		if err := c.refreshSubtreePaths(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

func sameParentID(a, b *uint) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
