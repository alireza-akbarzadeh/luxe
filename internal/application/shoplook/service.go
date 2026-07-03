package shoplook

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service serves storefront shop-the-look queries.
type Service struct {
	repo *postgres.ShopLookRepository
}

// NewService wires shop look use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: postgres.NewShopLookRepository(db)}
}

// List returns active shop looks for cards and homepage sections.
func (s *Service) List(ctx context.Context, limit int) (dto.ShopLookListResponse, error) {
	looks, err := s.repo.ListActive(ctx, limit)
	if err != nil {
		return dto.ShopLookListResponse{}, err
	}

	items := make([]dto.ShopLookListItem, 0, len(looks))
	for i := range looks {
		items = append(items, toListItem(&looks[i]))
	}
	return dto.ShopLookListResponse{Looks: items}, nil
}

// GetBySlug returns a shop look with tagged product cards.
func (s *Service) GetBySlug(ctx context.Context, slug string) (dto.ShopLookResponse, error) {
	look, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ShopLookResponse{}, utils.ErrNotFound("shop look not found")
		}
		return dto.ShopLookResponse{}, err
	}
	return toResponse(ctx, look), nil
}

func toListItem(look *models.ShopLook) dto.ShopLookListItem {
	if look == nil {
		return dto.ShopLookListItem{}
	}
	return dto.ShopLookListItem{
		ID:          look.ID,
		Slug:        look.Slug,
		Title:       look.Title,
		Description: look.Description,
		ImageURL:    look.ImageURL,
		TagCount:    len(look.Tags),
	}
}

func toResponse(ctx context.Context, look *models.ShopLook) dto.ShopLookResponse {
	if look == nil {
		return dto.ShopLookResponse{}
	}

	tags := make([]dto.ShopLookTagResponse, 0, len(look.Tags))
	for _, tag := range look.Tags {
		if tag.Product == nil {
			continue
		}
		tags = append(tags, dto.ShopLookTagResponse{
			ID:       tag.ID,
			XPercent: tag.XPercent,
			YPercent: tag.YPercent,
			Label:    tag.Label,
			Product: dto.HomeProductItem{
				ProductResponse: dto.ToProductResponse(ctx, *tag.Product),
			},
		})
	}

	return dto.ShopLookResponse{
		ID:          look.ID,
		Slug:        look.Slug,
		Title:       look.Title,
		Description: look.Description,
		ImageURL:    look.ImageURL,
		Tags:        tags,
	}
}
