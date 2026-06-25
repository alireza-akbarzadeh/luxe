package services

import (
	"context"

	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type StoreServiceInterface interface {
	ListStores(limit, offset int, filters dto.StoreFilter) ([]*models.Store, int64, error)
	GetByID(id uint) (*models.Store, error)
	GetBySlug(slug string) (*models.Store, error)
	Create(req dto.CreateStoreRequest) (*models.Store, error)
	Update(id uint, req dto.UpdateStoreRequest) (*models.Store, error)
	Delete(id uint) error
	FollowStore(userID, storeID uint) error
	UnfollowStore(userID, storeID uint) error
	IsFollowing(userID, storeID uint) (bool, error)
	GetFollowedStoreIDs(userID uint, storeIDs []uint) (map[uint]bool, error)
	ListStoreReviews(storeID uint, limit, offset int) ([]models.StoreReview, int64, dto.StoreReviewSummary, error)
	GetUserStoreReview(userID, storeID uint) (*models.StoreReview, error)
	CreateStoreReview(userID, storeID uint, req dto.CreateStoreReviewRequest) (*models.StoreReview, error)
	UpdateStoreReview(userID, reviewID uint, req dto.UpdateStoreReviewRequest) (*models.StoreReview, error)
	DeleteStoreReview(userID, reviewID uint) error
	ListVendorStores(ctx context.Context, userID uint, role string) ([]*models.Store, error)
}

type storeService struct {
	commands *appstore.Commands
	queries  *appstore.Queries
}

func NewStoreService(db *gorm.DB) StoreServiceInterface {
	repo := postgres.NewStoreRepository(db)
	queries := appstore.NewQueries(repo)
	return &storeService{
		commands: appstore.NewCommands(repo, queries),
		queries:  queries,
	}
}

func (s *storeService) ListStores(limit, offset int, filters dto.StoreFilter) ([]*models.Store, int64, error) {
	return s.queries.ListStores(limit, offset, filters)
}

func (s *storeService) GetByID(id uint) (*models.Store, error) {
	return s.queries.GetByID(id)
}

func (s *storeService) GetBySlug(slug string) (*models.Store, error) {
	return s.queries.GetBySlug(slug)
}

func (s *storeService) Create(req dto.CreateStoreRequest) (*models.Store, error) {
	return s.commands.Create(req)
}

func (s *storeService) Update(id uint, req dto.UpdateStoreRequest) (*models.Store, error) {
	return s.commands.Update(id, req)
}

func (s *storeService) Delete(id uint) error {
	return s.commands.Delete(id)
}

func (s *storeService) FollowStore(userID, storeID uint) error {
	return s.commands.FollowStore(userID, storeID)
}

func (s *storeService) UnfollowStore(userID, storeID uint) error {
	return s.commands.UnfollowStore(userID, storeID)
}

func (s *storeService) IsFollowing(userID, storeID uint) (bool, error) {
	return s.queries.IsFollowing(userID, storeID)
}

func (s *storeService) GetFollowedStoreIDs(userID uint, storeIDs []uint) (map[uint]bool, error) {
	return s.queries.GetFollowedStoreIDs(userID, storeIDs)
}

func (s *storeService) ListStoreReviews(storeID uint, limit, offset int) ([]models.StoreReview, int64, dto.StoreReviewSummary, error) {
	return s.queries.ListStoreReviews(storeID, limit, offset)
}

func (s *storeService) GetUserStoreReview(userID, storeID uint) (*models.StoreReview, error) {
	return s.queries.GetUserStoreReview(userID, storeID)
}

func (s *storeService) CreateStoreReview(userID, storeID uint, req dto.CreateStoreReviewRequest) (*models.StoreReview, error) {
	return s.commands.CreateStoreReview(userID, storeID, req)
}

func (s *storeService) UpdateStoreReview(userID, reviewID uint, req dto.UpdateStoreReviewRequest) (*models.StoreReview, error) {
	return s.commands.UpdateStoreReview(userID, reviewID, req)
}

func (s *storeService) DeleteStoreReview(userID, reviewID uint) error {
	return s.commands.DeleteStoreReview(userID, reviewID)
}

func (s *storeService) ListVendorStores(ctx context.Context, userID uint, role string) ([]*models.Store, error) {
	return s.queries.ListVendorStores(ctx, userID, role)
}
