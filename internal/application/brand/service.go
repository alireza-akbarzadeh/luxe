package brand

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	domainbrand "github.com/alireza-akbarzadeh/luxe/internal/domain/brand"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service handles brand HTTP-oriented use cases (DTO mapping and workflow sync).
type Service struct {
	engine   *workflow.Engine
	commands *Commands
	queries  *Queries
}

// NewService wires brand application use cases.
func NewService(db *gorm.DB, engine *workflow.Engine) *Service {
	repo := postgres.NewBrandRepository(db)
	return &Service{
		engine:   engine,
		commands: NewCommands(domainbrand.NewService(), repo),
		queries:  NewQueries(repo),
	}
}

func (s *Service) syncBrandWorkflow(ctx context.Context, brandID uint, status string) {
	if !appworkflow.ApplyBrandWorkflow(ctx, s.engine, brandID, status, nil) {
		utils.Log.WithField("brand_id", brandID).WithField("status", status).
			Debug("brand workflow sync skipped or failed")
	}
}

// Create creates a brand and returns an API response DTO.
func (s *Service) Create(ctx context.Context, req *dto.CreateBrandRequest) (*dto.BrandResponse, error) {
	brand, err := s.commands.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	s.syncBrandWorkflow(ctx, brand.ID, brand.Status)

	loaded, err := s.queries.GetByID(ctx, brand.ID)
	if err != nil {
		return nil, err
	}
	return ToResponse(loaded), nil
}

// GetByID returns a brand API response by ID.
func (s *Service) GetByID(ctx context.Context, id uint) (*dto.BrandResponse, error) {
	brand, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("brand not found")
		}
		return nil, err
	}
	return ToResponse(brand), nil
}

// List returns paginated brand API responses.
func (s *Service) List(ctx context.Context, req *dto.ListBrandsRequest) ([]dto.BrandResponse, int64, error) {
	items, total, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.BrandResponse, 0, len(items))
	for i := range items {
		resp = append(resp, *ToResponse(&items[i].Brand, items[i].ProductCount))
	}
	return resp, total, nil
}

// Update updates a brand and returns an API response DTO.
func (s *Service) Update(ctx context.Context, id uint, req *dto.UpdateBrandRequest) (*dto.BrandResponse, error) {
	brand, err := s.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("brand not found")
		}
		return nil, err
	}

	ApplyUpdateDTO(brand, req)

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
	return ToResponse(loaded), nil
}

// Delete removes a brand by ID.
func (s *Service) Delete(ctx context.Context, id uint) error {
	rows, err := s.commands.Delete(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return utils.ErrNotFound("brand not found")
	}
	return nil
}
