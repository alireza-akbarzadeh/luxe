package services

import (
	"context"

	appreview "github.com/alireza-akbarzadeh/luxe/internal/application/review"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type ReviewServiceInterface interface {
	Create(userID uint, req dto.CreateReviewRequest) (*models.Review, error)
	Update(userID, reviewID uint, req dto.UpdateReviewRequest) (*models.Review, error)
	Delete(userID, reviewID uint) error
	GetProductReviews(productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error)
	GetUserReviewForProduct(userID, productID uint) (*models.Review, error)
	ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, int64, error)
	PerformTransition(ctx context.Context, reviewID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error)
}

type reviewService struct {
	commands *appreview.Commands
	queries  *appreview.Queries
}

func NewReviewService(db *gorm.DB, engine *workflow.Engine) ReviewServiceInterface {
	repo := postgres.NewReviewRepository(db)
	return &reviewService{
		commands: appreview.NewCommands(repo, engine),
		queries:  appreview.NewQueries(repo),
	}
}

func (s *reviewService) Create(userID uint, req dto.CreateReviewRequest) (*models.Review, error) {
	return s.commands.Create(context.Background(), userID, req)
}

func (s *reviewService) Update(userID, reviewID uint, req dto.UpdateReviewRequest) (*models.Review, error) {
	return s.commands.Update(context.Background(), userID, reviewID, req)
}

func (s *reviewService) Delete(userID, reviewID uint) error {
	return s.commands.Delete(context.Background(), userID, reviewID)
}

func (s *reviewService) GetProductReviews(productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error) {
	return s.queries.GetProductReviews(context.Background(), productID, limit, offset)
}

func (s *reviewService) GetUserReviewForProduct(userID, productID uint) (*models.Review, error) {
	return s.queries.GetUserReviewForProduct(context.Background(), userID, productID)
}

func (s *reviewService) ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, int64, error) {
	return s.queries.ListAdmin(ctx, filters)
}

func (s *reviewService) PerformTransition(ctx context.Context, reviewID uint, event, note, actorRole string, actorID *uint) (*workflow.TransitionResult, error) {
	return s.commands.PerformTransition(ctx, reviewID, event, note, actorRole, actorID)
}
