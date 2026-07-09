package category

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	domaincategory "github.com/alireza-akbarzadeh/luxe/internal/domain/category"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service handles category HTTP-oriented use cases (slug, workflow sync, error mapping).
type Service struct {
	engine   *workflow.Engine
	commands *Commands
	queries  *Queries
}

// NewService wires category application use cases.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewCategoryRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domaincategory.NewService(), repo),
		queries:  NewQueries(repo),
	}
}

func (s *Service) syncCategoryWorkflow(ctx context.Context, categoryID uint, isActive bool) {
	if !appworkflow.ApplyCategoryWorkflow(ctx, s.engine, categoryID, isActive, nil) {
		utils.Log.WithField("category_id", categoryID).Debug("category workflow sync skipped or failed")
	}
}

// Create inserts a category and reloads it with relations.
func (s *Service) Create(req dto.CreateCategoryRequest) (*models.Category, error) {
	ctx := context.Background()
	slug := req.Slug
	if slug == "" {
		slug = GenerateSlug(req.Name)
	}
	unique, err := s.commands.UniqueSlug(ctx, slug, 0)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	category, err := s.commands.Create(ctx, req, unique)
	if err != nil {
		if errors.Is(err, domaincategory.ErrInvalidName) {
			return nil, utils.ErrValidationFailed(err.Error())
		}
		return nil, utils.ErrInternal(err)
	}
	s.syncCategoryWorkflow(ctx, category.ID, category.IsActive)
	return s.GetByID(category.ID)
}

// GetByID returns a category by ID.
func (s *Service) GetByID(id uint) (*models.Category, error) {
	category, err := s.queries.GetByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("category not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return category, nil
}

// GetBySlug returns a category by slug.
func (s *Service) GetBySlug(slug string) (*models.Category, error) {
	category, err := s.queries.GetBySlug(context.Background(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("category not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return category, nil
}

// Update updates a category and reloads it.
func (s *Service) Update(id uint, req dto.UpdateCategoryRequest) (*models.Category, error) {
	ctx := context.Background()
	category, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	newSlug := ""
	if req.Name != nil {
		baseSlug := GenerateSlug(*req.Name)
		if req.Slug == nil || *req.Slug == "" {
			unique, err := s.commands.UniqueSlug(ctx, baseSlug, id)
			if err != nil {
				return nil, utils.ErrInternal(err)
			}
			newSlug = unique
		}
	}
	if req.Slug != nil && *req.Slug != "" {
		unique, err := s.commands.UniqueSlug(ctx, *req.Slug, id)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		newSlug = unique
	}

	ApplyUpdateDTO(category, req, newSlug)
	if req.ParentID != nil {
		if err := s.commands.UpdateLevelAndPath(ctx, category); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	if err := s.commands.Save(ctx, category, id); err != nil {
		return nil, utils.ErrInternal(err)
	}
	if req.IsActive != nil {
		s.syncCategoryWorkflow(ctx, id, *req.IsActive)
	}
	return s.GetByID(id)
}

// Delete removes a category by ID.
func (s *Service) Delete(id uint) error {
	rows, err := s.commands.Delete(context.Background(), id)
	if err != nil {
		if errors.Is(err, ErrHasChildren) {
			return utils.ErrBadRequest("cannot delete category with children; delete children first")
		}
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("category not found")
	}
	return nil
}

// List returns categories matching filters.
func (s *Service) List(filters dto.CategoryListFilters) ([]models.Category, int64, error) {
	categories, total, err := s.queries.List(context.Background(), filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return categories, total, nil
}

// BulkCreate creates multiple categories.
func (s *Service) BulkCreate(categories []dto.CreateCategoryRequest) ([]*models.Category, error) {
	if len(categories) == 0 {
		return nil, utils.ErrBadRequest("no categories provided")
	}
	created, err := s.commands.BulkCreate(context.Background(), categories, s.commands.UniqueSlug)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return created, nil
}

// BulkDelete removes multiple categories by ID.
func (s *Service) BulkDelete(ids []uint) error {
	if len(ids) == 0 {
		return utils.ErrBadRequest("no category IDs provided")
	}
	rows, err := s.commands.BulkDelete(context.Background(), ids)
	if err != nil {
		if errors.Is(err, ErrHasChildren) {
			return utils.ErrBadRequest("cannot delete categories that have children; delete children first")
		}
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("no categories found to delete")
	}
	return nil
}

// Reorder updates category sort order and optional parent assignments.
func (s *Service) Reorder(req dto.ReorderCategoriesRequest) error {
	if len(req.Items) == 0 {
		return utils.ErrBadRequest("no categories provided")
	}
	err := s.commands.Reorder(context.Background(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidParentMove) {
			return utils.ErrBadRequest("cannot move a category under itself or its descendants")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("one or more categories were not found")
		}
		return utils.ErrInternal(err)
	}
	return nil
}
