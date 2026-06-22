package services

import (
	"context"
	"errors"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type ReviewServiceInterface interface {
	Create(userID uint, req dto.CreateReviewRequest) (*models.Review, error)
	Update(userID, reviewID uint, req dto.UpdateReviewRequest) (*models.Review, error)
	Delete(userID, reviewID uint) error
	GetProductReviews(productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error)
	GetUserReviewForProduct(userID, productID uint) (*models.Review, error)
	ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, int64, error)
	Moderate(ctx context.Context, reviewID uint, status string) (*models.Review, error)
}

type reviewService struct {
	db *gorm.DB
}

func NewReviewService(db *gorm.DB) ReviewServiceInterface {
	return &reviewService{db: db}
}

func approvedReviewScope(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", constants.ReviewStatusApproved)
}

// Create review with pending moderation status.
func (s *reviewService) Create(userID uint, req dto.CreateReviewRequest) (*models.Review, error) {
	var existing models.Review
	err := s.db.Where("user_id = ? AND product_id = ?", userID, req.ProductID).First(&existing).Error
	if err == nil {
		return nil, utils.ErrBadRequest("you have already reviewed this product")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	review := &models.Review{
		ProductID: req.ProductID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		Title:     req.Title,
		Status:    constants.ReviewStatusPending,
	}
	if err := s.db.Create(review).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return review, nil
}

// Update review; resets moderation when content changes.
func (s *reviewService) Update(userID, reviewID uint, req dto.UpdateReviewRequest) (*models.Review, error) {
	var review models.Review
	err := s.db.Preload("User").Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("review not found")
		}
		return nil, utils.ErrInternal(err)
	}

	previousStatus := review.Status
	contentChanged := false

	if req.Rating != nil {
		review.Rating = *req.Rating
		contentChanged = true
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
		contentChanged = true
	}
	if req.Title != nil {
		review.Title = *req.Title
		contentChanged = true
	}

	if contentChanged && review.Status != constants.ReviewStatusPending {
		review.Status = constants.ReviewStatusPending
	}

	if err := s.db.Save(&review).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if previousStatus == constants.ReviewStatusApproved || review.Status == constants.ReviewStatusApproved {
		s.updateProductStats(review.ProductID)
	}

	return &review, nil
}

// Delete review and recalc product stats when an approved review is removed.
func (s *reviewService) Delete(userID, reviewID uint) error {
	var review models.Review
	err := s.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("review not found")
		}
		return utils.ErrInternal(err)
	}

	wasApproved := review.Status == constants.ReviewStatusApproved
	productID := review.ProductID

	if err := s.db.Delete(&review).Error; err != nil {
		return utils.ErrInternal(err)
	}

	if wasApproved {
		s.updateProductStats(productID)
	}
	return nil
}

func (s *reviewService) updateProductStats(productID uint) {
	var result struct {
		AvgRating float32
		Count     int
	}
	s.db.Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Scan(&result)

	s.db.Model(&models.Product{}).Where("id = ?", productID).
		Updates(map[string]interface{}{
			"rating":        result.AvgRating,
			"reviews_count": result.Count,
		})
}

func (s *reviewService) buildReviewSummary(productID uint) (dto.ReviewSummary, error) {
	summary := dto.ReviewSummary{
		Counts: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
	}

	var result struct {
		AvgRating float64
		Count     int64
	}
	if err := s.db.Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Scan(&result).Error; err != nil {
		return summary, utils.ErrInternal(err)
	}

	summary.Average = result.AvgRating
	summary.Total = result.Count

	type ratingCount struct {
		Rating int
		Count  int
	}
	var rows []ratingCount
	if err := s.db.Model(&models.Review{}).
		Select("rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Group("rating").
		Scan(&rows).Error; err != nil {
		return summary, utils.ErrInternal(err)
	}

	for _, row := range rows {
		summary.Counts[strconv.Itoa(row.Rating)] = row.Count
	}

	return summary, nil
}

// GetProductReviews returns paginated approved reviews and rating summary for a product.
func (s *reviewService) GetProductReviews(productID uint, limit, offset int) ([]models.Review, int64, dto.ReviewSummary, error) {
	summary, err := s.buildReviewSummary(productID)
	if err != nil {
		return nil, 0, summary, err
	}

	var reviews []models.Review
	query := s.db.Model(&models.Review{}).
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope)

	if err := query.Preload("User").Order("created_at DESC").Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		return nil, 0, summary, utils.ErrInternal(err)
	}

	return reviews, summary.Total, summary, nil
}

// GetUserReviewForProduct returns the authenticated user's review for a product (if any).
func (s *reviewService) GetUserReviewForProduct(userID, productID uint) (*models.Review, error) {
	var review models.Review
	err := s.db.Preload("User").Where("user_id = ? AND product_id = ?", userID, productID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return &review, nil
}

// ListAdmin returns paginated reviews for moderation.
func (s *reviewService) ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, int64, error) {
	query := s.db.WithContext(ctx).Model(&models.Review{})

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.ProductID > 0 {
		query = query.Where("product_id = ?", filters.ProductID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var reviews []models.Review
	if err := query.
		Preload("User").
		Preload("Product").
		Order("created_at DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&reviews).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return reviews, total, nil
}

// Moderate approves or rejects a review and recalculates product stats when needed.
func (s *reviewService) Moderate(ctx context.Context, reviewID uint, status string) (*models.Review, error) {
	if status != constants.ReviewStatusApproved && status != constants.ReviewStatusRejected {
		return nil, utils.ErrBadRequest("invalid review status")
	}

	var review models.Review
	err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Product").
		Where("id = ?", reviewID).
		First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("review not found")
		}
		return nil, utils.ErrInternal(err)
	}

	previousStatus := review.Status
	if previousStatus == status {
		return &review, nil
	}

	review.Status = status
	if err := s.db.WithContext(ctx).Save(&review).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if previousStatus == constants.ReviewStatusApproved || status == constants.ReviewStatusApproved {
		s.updateProductStats(review.ProductID)
	}

	return &review, nil
}
