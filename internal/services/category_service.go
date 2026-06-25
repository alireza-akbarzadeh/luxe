package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appcategory "github.com/alireza-akbarzadeh/luxe/internal/application/category"
	domaincategory "github.com/alireza-akbarzadeh/luxe/internal/domain/category"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

type CategoryServiceInterface interface {
	Create(req dto.CreateCategoryRequest) (*models.Category, error)
	GetByID(id uint) (*models.Category, error)
	GetBySlug(slug string) (*models.Category, error)
	Update(id uint, req dto.UpdateCategoryRequest) (*models.Category, error)
	Delete(id uint) error
	List(filters dto.CategoryListFilters) ([]models.Category, int64, error)
	BulkCreate(categories []dto.CreateCategoryRequest) ([]*models.Category, error)
	BulkDelete(ids []uint) error
}

type categoryService struct {
	engine   *workflow.Engine
	commands *appcategory.Commands
	queries  *appcategory.Queries
}

type BulkDeleteCategoryRequest struct {
	IDs []uint `json:"ids" validate:"required,min=1"`
}

func NewCategoryService(db *gorm.DB, engine *workflow.Engine) CategoryServiceInterface {
	repo := postgres.NewCategoryRepository(db)
	return &categoryService{
		engine:   engine,
		commands: appcategory.NewCommands(domaincategory.NewService(), repo),
		queries:  appcategory.NewQueries(repo),
	}
}

func (s *categoryService) syncCategoryWorkflow(ctx context.Context, categoryID uint, isActive bool) {
	if !applyCategoryWorkflow(ctx, s.engine, categoryID, isActive, nil) {
		utils.Log.WithField("category_id", categoryID).Debug("category workflow sync skipped or failed")
	}
}

func (s *categoryService) Create(req dto.CreateCategoryRequest) (*models.Category, error) {
	ctx := context.Background()
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
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

func (s *categoryService) GetByID(id uint) (*models.Category, error) {
	category, err := s.queries.GetByID(context.Background(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("category not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return category, nil
}

func (s *categoryService) GetBySlug(slug string) (*models.Category, error) {
	category, err := s.queries.GetBySlug(context.Background(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("category not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return category, nil
}

func (s *categoryService) Update(id uint, req dto.UpdateCategoryRequest) (*models.Category, error) {
	ctx := context.Background()
	category, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	newSlug := ""
	if req.Name != nil {
		baseSlug := generateSlug(*req.Name)
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

	appcategory.ApplyUpdateDTO(category, req, newSlug)
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

func (s *categoryService) Delete(id uint) error {
	rows, err := s.commands.Delete(context.Background(), id)
	if err != nil {
		if errors.Is(err, appcategory.ErrHasChildren) {
			return utils.ErrBadRequest("cannot delete category with children; delete children first")
		}
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("category not found")
	}
	return nil
}

func (s *categoryService) List(filters dto.CategoryListFilters) ([]models.Category, int64, error) {
	categories, total, err := s.queries.List(context.Background(), filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return categories, total, nil
}

func (s *categoryService) BulkCreate(categories []dto.CreateCategoryRequest) ([]*models.Category, error) {
	if len(categories) == 0 {
		return nil, utils.ErrBadRequest("no categories provided")
	}
	created, err := s.commands.BulkCreate(context.Background(), categories, s.commands.UniqueSlug)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return created, nil
}

func (s *categoryService) BulkDelete(ids []uint) error {
	if len(ids) == 0 {
		return utils.ErrBadRequest("no category IDs provided")
	}
	rows, err := s.commands.BulkDelete(context.Background(), ids)
	if err != nil {
		if errors.Is(err, appcategory.ErrHasChildren) {
			return utils.ErrBadRequest("cannot delete categories that have children; delete children first")
		}
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("no categories found to delete")
	}
	return nil
}
