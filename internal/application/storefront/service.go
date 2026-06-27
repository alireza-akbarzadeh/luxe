package storefront

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Service orchestrates storefront landing-page queries.
type Service struct {
	repo *postgres.StorefrontRepository
}

// NewService wires storefront use cases.
func NewService(repo *postgres.StorefrontRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListDiscountedProducts(ctx context.Context, limit, offset int) ([]*models.Product, int64, error) {
	products, total, err := s.repo.ListDiscountedProducts(ctx, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return products, total, nil
}

func (s *Service) ListBestSellers(ctx context.Context, limit int) ([]*models.Product, map[uint]int64, error) {
	if limit <= 0 {
		limit = 8
	}
	if limit > 24 {
		limit = 24
	}

	rows, err := s.repo.ListBestSellerIDs(ctx, limit)
	if err != nil {
		return nil, nil, utils.ErrInternal(err)
	}
	if len(rows) == 0 {
		products, err := s.repo.ListTopRatedProducts(ctx, limit)
		if err != nil {
			return nil, nil, utils.ErrInternal(err)
		}
		return products, map[uint]int64{}, nil
	}

	ids := make([]uint, len(rows))
	unitsSold := make(map[uint]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ProductID
		unitsSold[row.ProductID] = row.UnitsSold
	}

	products, err := s.repo.ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, nil, utils.ErrInternal(err)
	}
	return products, unitsSold, nil
}

func (s *Service) ListPopularBrands(ctx context.Context, limit int) ([]dto.StorefrontBrandItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 24 {
		limit = 24
	}

	rows, err := s.repo.ListPopularBrands(ctx, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	items := make([]dto.StorefrontBrandItem, 0, len(rows))
	for _, row := range rows {
		logo := ""
		if row.LogoURL != nil {
			logo = *row.LogoURL
		}
		items = append(items, dto.StorefrontBrandItem{
			ID:           row.ID,
			Name:         row.Name,
			Slug:         row.Slug,
			LogoURL:      logo,
			ProductCount: row.ProductCount,
		})
	}
	return items, nil
}

func (s *Service) ListHomeCategories(ctx context.Context, limit int) ([]dto.StorefrontCategoryItem, error) {
	if limit <= 0 {
		limit = 8
	}
	if limit > 16 {
		limit = 16
	}

	rows, err := s.repo.ListHomeCategories(ctx, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if len(rows) == 0 {
		return []dto.StorefrontCategoryItem{}, nil
	}

	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	covers, err := s.repo.FindCategoryCoverImages(ctx, ids)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	items := make([]dto.StorefrontCategoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.StorefrontCategoryItem{
			ID:           row.ID,
			Name:         row.Name,
			Slug:         row.Slug,
			Description:  row.Description,
			ImageURL:     covers[row.ID],
			ProductCount: row.ProductCount,
		})
	}
	return items, nil
}
