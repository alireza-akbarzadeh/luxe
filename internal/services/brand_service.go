package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appbrand "github.com/alireza-akbarzadeh/luxe/internal/application/brand"
	domainbrand "github.com/alireza-akbarzadeh/luxe/internal/domain/brand"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

type BrandServiceInterface interface {
	Create(ctx context.Context, req *dto.CreateBrandRequest) (*dto.BrandResponse, error)
	GetByID(ctx context.Context, id uint) (*dto.BrandResponse, error)
	List(ctx context.Context, req *dto.ListBrandsRequest) ([]dto.BrandResponse, int64, error)
	Update(ctx context.Context, id uint, req *dto.UpdateBrandRequest) (*dto.BrandResponse, error)
	Delete(ctx context.Context, id uint) error
}

type brandService struct {
	engine   *workflow.Engine
	commands *appbrand.Commands
	queries  *appbrand.Queries
}

func NewBrandService(db *gorm.DB, engine *workflow.Engine) BrandServiceInterface {
	repo := postgres.NewBrandRepository(db)
	return &brandService{
		engine:   engine,
		commands: appbrand.NewCommands(domainbrand.NewService(), repo),
		queries:  appbrand.NewQueries(repo),
	}
}

func (s *brandService) syncBrandWorkflow(ctx context.Context, brandID uint, status string) {
	if !applyBrandWorkflow(ctx, s.engine, brandID, status, nil) {
		utils.Log.WithField("brand_id", brandID).WithField("status", status).
			Debug("brand workflow sync skipped or failed")
	}
}

func (s *brandService) Create(ctx context.Context, req *dto.CreateBrandRequest) (*dto.BrandResponse, error) {
	brand, err := s.commands.Create(ctx, req)
	if err != nil {
		if errors.Is(err, domainbrand.ErrInvalidName) {
			return nil, err
		}
		return nil, err
	}

	s.syncBrandWorkflow(ctx, brand.ID, brand.Status)

	loaded, err := s.queries.GetByID(ctx, brand.ID)
	if err != nil {
		return nil, err
	}
	return appbrand.ToResponse(loaded), nil
}

func (s *brandService) GetByID(ctx context.Context, id uint) (*dto.BrandResponse, error) {
	brand, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return appbrand.ToResponse(brand), nil
}

func (s *brandService) List(ctx context.Context, req *dto.ListBrandsRequest) ([]dto.BrandResponse, int64, error) {
	brands, total, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.BrandResponse, 0, len(brands))
	for i := range brands {
		resp = append(resp, *appbrand.ToResponse(&brands[i]))
	}
	return resp, total, nil
}

func (s *brandService) Update(ctx context.Context, id uint, req *dto.UpdateBrandRequest) (*dto.BrandResponse, error) {
	brand, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	appbrand.ApplyUpdateDTO(brand, req)

	if err := s.commands.Update(ctx, brand); err != nil {
		return nil, err
	}

	if req.Status != nil {
		s.syncBrandWorkflow(ctx, id, *req.Status)
	}

	loaded, err := s.queries.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return appbrand.ToResponse(loaded), nil
}

func (s *brandService) Delete(ctx context.Context, id uint) error {
	rows, err := s.commands.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

var ErrNotFound = errors.New("resource not found")
