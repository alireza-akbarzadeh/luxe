package services

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
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
	db *gorm.DB
}

func NewBrandService(db *gorm.DB) BrandServiceInterface {
	return &brandService{db: db}
}

func (s *brandService) Create(ctx context.Context, req *dto.CreateBrandRequest) (*dto.BrandResponse, error) {
	status := "draft"
	if req.Status != nil {
		status = *req.Status
	}

	brand := models.Brand{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		Status:      status,
	}

	if err := s.db.WithContext(ctx).Create(&brand).Error; err != nil {
		return nil, err
	}

	return brandToResponse(&brand), nil
}

func (s *brandService) GetByID(ctx context.Context, id uint) (*dto.BrandResponse, error) {
	var brand models.Brand
	if err := s.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return brandToResponse(&brand), nil
}

func (s *brandService) List(ctx context.Context, req *dto.ListBrandsRequest) ([]dto.BrandResponse, int64, error) {
	var brands []models.Brand
	var total int64

	query := s.db.WithContext(ctx).Model(&models.Brand{})

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", search, search)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(offset).Limit(req.Limit).Order("created_at DESC").Find(&brands).Error; err != nil {
		return nil, 0, err
	}

	var resp []dto.BrandResponse
	for _, b := range brands {
		resp = append(resp, *brandToResponse(&b))
	}
	return resp, total, nil
}

func (s *brandService) Update(ctx context.Context, id uint, req *dto.UpdateBrandRequest) (*dto.BrandResponse, error) {
	var brand models.Brand
	if err := s.db.WithContext(ctx).First(&brand, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Apply updates only to non-nil fields
	if req.Name != nil {
		brand.Name = *req.Name
	}
	if req.Slug != nil {
		brand.Slug = *req.Slug
	}
	if req.Description != nil {
		brand.Description = req.Description
	}
	if req.LogoURL != nil {
		brand.LogoURL = req.LogoURL
	}
	if req.Status != nil {
		brand.Status = *req.Status
	}

	if err := s.db.WithContext(ctx).Save(&brand).Error; err != nil {
		return nil, err
	}

	return brandToResponse(&brand), nil
}

func (s *brandService) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.Brand{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- helpers ----------
var ErrNotFound = errors.New("resource not found")

func brandToResponse(b *models.Brand) *dto.BrandResponse {
	return &dto.BrandResponse{
		ID:          b.ID,
		Name:        b.Name,
		Slug:        b.Slug,
		Description: b.Description,
		LogoURL:     b.LogoURL,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}
