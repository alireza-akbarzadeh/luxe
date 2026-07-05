package communityshoppinglist

import (
	"context"
	"errors"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service serves public community shopping list queries.
type Service struct {
	repo *postgres.CommunityShoppingListRepository
}

// NewService wires community shopping list use cases.
func NewService(db *gorm.DB) *Service {
	return &Service{repo: postgres.NewCommunityShoppingListRepository(db)}
}

// List returns active community shopping lists for discovery pages.
func (s *Service) List(ctx context.Context, limit int) (dto.CommunityShoppingListListResponse, error) {
	lists, err := s.repo.ListActive(ctx, limit)
	if err != nil {
		return dto.CommunityShoppingListListResponse{}, err
	}

	items := make([]dto.CommunityShoppingListListItem, 0, len(lists))
	for i := range lists {
		items = append(items, toListItem(&lists[i]))
	}
	return dto.CommunityShoppingListListResponse{Lists: items}, nil
}

// GetBySlug returns a community shopping list with product items.
func (s *Service) GetBySlug(ctx context.Context, slug string) (dto.CommunityShoppingListResponse, error) {
	list, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CommunityShoppingListResponse{}, utils.ErrNotFound("shopping list not found")
		}
		return dto.CommunityShoppingListResponse{}, err
	}
	return toResponse(ctx, list), nil
}

func toListItem(list *models.CommunityShoppingList) dto.CommunityShoppingListListItem {
	if list == nil {
		return dto.CommunityShoppingListListItem{}
	}
	return dto.CommunityShoppingListListItem{
		ID:            list.ID,
		Slug:          list.Slug,
		Title:         list.Title,
		Description:   list.Description,
		Theme:         list.Theme,
		CoverImageURL: list.CoverImageURL,
		AuthorName:    list.AuthorName,
		AuthorHandle:  list.AuthorHandle,
		ItemCount:     len(list.Items),
	}
}

func toResponse(ctx context.Context, list *models.CommunityShoppingList) dto.CommunityShoppingListResponse {
	if list == nil {
		return dto.CommunityShoppingListResponse{}
	}

	items := make([]dto.CommunityShoppingListItemResponse, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Product == nil {
			continue
		}
		items = append(items, dto.CommunityShoppingListItemResponse{
			ID:   item.ID,
			Note: item.Note,
			Product: dto.HomeProductItem{
				ProductResponse: dto.ToProductResponse(ctx, *item.Product),
			},
		})
	}

	return dto.CommunityShoppingListResponse{
		ID:            list.ID,
		Slug:          list.Slug,
		Title:         list.Title,
		Description:   list.Description,
		Theme:         list.Theme,
		CoverImageURL: list.CoverImageURL,
		AuthorName:    list.AuthorName,
		AuthorHandle:  list.AuthorHandle,
		Items:         items,
	}
}
