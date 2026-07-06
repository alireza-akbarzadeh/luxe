package publiccollection

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service serves public community collection queries.
type Service struct {
	repo *postgres.PublicCollectionRepository
}

// NewService wires public collection use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: postgres.NewPublicCollectionRepository(db)}
}

// List returns active public collections for discovery pages.
func (s *Service) List(ctx context.Context, limit int) (dto.PublicCollectionListResponse, error) {
	collections, err := s.repo.ListActive(ctx, limit)
	if err != nil {
		return dto.PublicCollectionListResponse{}, err
	}

	items := make([]dto.PublicCollectionListItem, 0, len(collections))
	for i := range collections {
		items = append(items, toListItem(&collections[i]))
	}
	return dto.PublicCollectionListResponse{Collections: items}, nil
}

// GetBySlug returns a public collection with product items.
func (s *Service) GetBySlug(ctx context.Context, slug string) (dto.PublicCollectionResponse, error) {
	collection, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PublicCollectionResponse{}, utils.ErrNotFound("public collection not found")
		}
		return dto.PublicCollectionResponse{}, err
	}
	return toResponse(ctx, collection), nil
}

func toListItem(collection *models.PublicCollection) dto.PublicCollectionListItem {
	if collection == nil {
		return dto.PublicCollectionListItem{}
	}
	return dto.PublicCollectionListItem{
		ID:            collection.ID,
		Slug:          collection.Slug,
		Title:         collection.Title,
		Description:   collection.Description,
		Theme:         collection.Theme,
		CoverImageURL: collection.CoverImageURL,
		AuthorName:    collection.AuthorName,
		AuthorHandle:  collection.AuthorHandle,
		ItemCount:     len(collection.Items),
	}
}

func toResponse(ctx context.Context, collection *models.PublicCollection) dto.PublicCollectionResponse {
	if collection == nil {
		return dto.PublicCollectionResponse{}
	}

	items := make([]dto.PublicCollectionItemResponse, 0, len(collection.Items))
	for _, item := range collection.Items {
		if item.Product == nil {
			continue
		}
		items = append(items, dto.PublicCollectionItemResponse{
			ID:   item.ID,
			Note: item.Note,
			Product: dto.HomeProductItem{
				ProductResponse: dto.ToProductResponse(ctx, *item.Product),
			},
		})
	}

	return dto.PublicCollectionResponse{
		ID:            collection.ID,
		Slug:          collection.Slug,
		Title:         collection.Title,
		Description:   collection.Description,
		Theme:         collection.Theme,
		CoverImageURL: collection.CoverImageURL,
		AuthorName:    collection.AuthorName,
		AuthorHandle:  collection.AuthorHandle,
		Items:         items,
	}
}
