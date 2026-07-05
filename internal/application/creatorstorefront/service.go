package creatorstorefront

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service serves public creator storefront queries.
type Service struct {
	repo *postgres.CreatorStorefrontRepository
}

// NewService wires creator storefront use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: postgres.NewCreatorStorefrontRepository(db)}
}

// List returns active creator profiles for discovery pages.
func (s *Service) List(ctx context.Context, limit int) (dto.CreatorStorefrontListResponse, error) {
	creators, err := s.repo.ListActive(ctx, limit)
	if err != nil {
		return dto.CreatorStorefrontListResponse{}, err
	}

	items := make([]dto.CreatorStorefrontListItem, 0, len(creators))
	for i := range creators {
		items = append(items, toListItem(&creators[i]))
	}
	return dto.CreatorStorefrontListResponse{Creators: items}, nil
}

// GetBySlug returns a creator profile with curated product picks.
func (s *Service) GetBySlug(ctx context.Context, slug string) (dto.CreatorStorefrontResponse, error) {
	creator, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CreatorStorefrontResponse{}, utils.ErrNotFound("creator not found")
		}
		return dto.CreatorStorefrontResponse{}, err
	}
	return toResponse(ctx, creator), nil
}

func toListItem(creator *models.Creator) dto.CreatorStorefrontListItem {
	if creator == nil {
		return dto.CreatorStorefrontListItem{}
	}
	return dto.CreatorStorefrontListItem{
		ID:            creator.ID,
		Slug:          creator.Slug,
		DisplayName:   creator.DisplayName,
		Handle:        creator.Handle,
		Bio:           creator.Bio,
		Specialty:     creator.Specialty,
		AvatarURL:     creator.AvatarURL,
		CoverImageURL: creator.CoverImageURL,
		PickCount:     len(creator.Picks),
	}
}

func toResponse(ctx context.Context, creator *models.Creator) dto.CreatorStorefrontResponse {
	if creator == nil {
		return dto.CreatorStorefrontResponse{}
	}

	picks := make([]dto.CreatorPickResponse, 0, len(creator.Picks))
	for _, pick := range creator.Picks {
		if pick.Product == nil {
			continue
		}
		picks = append(picks, dto.CreatorPickResponse{
			ID:       pick.ID,
			Headline: pick.Headline,
			Product: dto.HomeProductItem{
				ProductResponse: dto.ToProductResponse(ctx, *pick.Product),
			},
		})
	}

	return dto.CreatorStorefrontResponse{
		ID:            creator.ID,
		Slug:          creator.Slug,
		DisplayName:   creator.DisplayName,
		Handle:        creator.Handle,
		Bio:           creator.Bio,
		Specialty:     creator.Specialty,
		AvatarURL:     creator.AvatarURL,
		CoverImageURL: creator.CoverImageURL,
		InstagramURL:  creator.InstagramURL,
		Picks:         picks,
	}
}
