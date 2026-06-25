package review

import (
	"context"
	"errors"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates review read use cases.
type Queries struct {
	repo *postgres.ReviewRepository
}

// NewQueries creates review query use cases.
func NewQueries(repo *postgres.ReviewRepository) *Queries {
	return &Queries{repo: repo}
}

// buildReviewSummary aggregates approved review stats for a product.
func (q *Queries) buildReviewSummary(ctx context.Context, productID uint) (dto.ReviewSummary, error) {
	summary := dto.ReviewSummary{
		Counts: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
	}

	avg, total, err := q.repo.SummaryStats(ctx, productID)
	if err != nil {
		return summary, utils.ErrInternal(err)
	}
	summary.Average = avg
	summary.Total = total

	rows, err := q.repo.RatingCounts(ctx, productID)
	if err != nil {
		return summary, utils.ErrInternal(err)
	}
	for rating, count := range rows {
		summary.Counts[strconv.Itoa(rating)] = count
	}
	return summary, nil
}

// GetProductReviews returns paginated approved reviews and summary.
func (q *Queries) GetProductReviews(ctx context.Context, productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error) {
	summary, err := q.buildReviewSummary(ctx, productID)
	if err != nil {
		return nil, 0, summary, err
	}

	reviews, err := q.repo.ListApproved(ctx, productID, limit, offset)
	if err != nil {
		return nil, 0, summary, utils.ErrInternal(err)
	}
	return reviews, summary.Total, summary, nil
}

// GetUserReviewForProduct returns the user's review for a product if present.
func (q *Queries) GetUserReviewForProduct(ctx context.Context, userID, productID uint) (*models.Review, error) {
	review, err := q.repo.FindUserReview(ctx, userID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return review, nil
}

// ListAdmin returns paginated reviews for moderation.
func (q *Queries) ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, int64, error) {
	total, err := q.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	reviews, err := q.repo.ListAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return reviews, total, nil
}
