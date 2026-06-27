package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ReviewRepository persists product reviews with GORM.
type ReviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository creates a GORM-backed review repository.
func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func approvedReviewScope(db *gorm.DB) *gorm.DB {
	return db.Where(`reviews.workflow_state_id IN (
		SELECT ws.id FROM workflow_states ws
		INNER JOIN workflows w ON w.id = ws.workflow_id
		WHERE w.key = ? AND ws.code = 'approved'
	)`, constants.WorkflowEntityReview)
}

// FindByUserAndProduct loads a review for a user/product pair.
func (r *ReviewRepository) FindByUserAndProduct(ctx context.Context, userID, productID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByID loads a review with relations.
func (r *ReviewRepository) FindByID(ctx context.Context, id uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Product").Preload("WorkflowState").
		First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByIDForUser loads a review owned by a user.
func (r *ReviewRepository) FindByIDForUser(ctx context.Context, reviewID, userID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindByIDFields loads selected review fields.
func (r *ReviewRepository) FindByIDFields(ctx context.Context, reviewID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).Select("product_id", "status").First(&review, reviewID).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// Create inserts a review.
func (r *ReviewRepository) Create(ctx context.Context, review *models.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

// Save persists review changes.
func (r *ReviewRepository) Save(ctx context.Context, review *models.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

// Delete removes a review.
func (r *ReviewRepository) Delete(ctx context.Context, review *models.Review) error {
	return r.db.WithContext(ctx).Delete(review).Error
}

// ProductStats scans approved review aggregates for a product.
func (r *ReviewRepository) ProductStats(ctx context.Context, productID uint) (avgRating float32, count int, err error) {
	var result struct {
		AvgRating float32
		Count     int
	}
	err = r.db.WithContext(ctx).Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Scan(&result).Error
	return result.AvgRating, result.Count, err
}

// UpdateProductRating updates product rating fields.
func (r *ReviewRepository) UpdateProductRating(ctx context.Context, productID uint, rating float32, reviewsCount int) error {
	return r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).
		Updates(map[string]interface{}{
			"rating":        rating,
			"reviews_count": reviewsCount,
		}).Error
}

// SummaryStats returns average rating and total for approved reviews.
func (r *ReviewRepository) SummaryStats(ctx context.Context, productID uint) (avg float64, total int64, err error) {
	var result struct {
		AvgRating float64
		Count     int64
	}
	err = r.db.WithContext(ctx).Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Scan(&result).Error
	return result.AvgRating, result.Count, err
}

// RatingCounts returns per-rating counts for approved reviews.
func (r *ReviewRepository) RatingCounts(ctx context.Context, productID uint) (map[int]int, error) {
	type ratingCount struct {
		Rating int
		Count  int
	}
	var rows []ratingCount
	err := r.db.WithContext(ctx).Model(&models.Review{}).
		Select("rating, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Group("rating").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int]int, len(rows))
	for _, row := range rows {
		out[row.Rating] = row.Count
	}
	return out, nil
}

// ListApproved returns paginated approved reviews for a product.
func (r *ReviewRepository) ListApproved(ctx context.Context, productID uint, limit, offset int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.WithContext(ctx).Model(&models.Review{}).
		Where("product_id = ?", productID).
		Scopes(approvedReviewScope).
		Preload("User").Preload("WorkflowState").
		Order("created_at DESC").Limit(limit).Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

// FindUserReview loads a user's review for a product.
func (r *ReviewRepository) FindUserReview(ctx context.Context, userID, productID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).Preload("User").Preload("WorkflowState").
		Where("user_id = ? AND product_id = ?", userID, productID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) applyAdminFilters(q *gorm.DB, filters dto.AdminReviewListFilters) *gorm.DB {
	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.ProductID > 0 {
		q = q.Where("product_id = ?", filters.ProductID)
	}
	return q
}

// CountAdmin counts reviews matching admin filters.
func (r *ReviewRepository) CountAdmin(ctx context.Context, filters dto.AdminReviewListFilters) (int64, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Review{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// CountByUser counts reviews authored by a user.
func (r *ReviewRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Review{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

// ListByUser returns paginated reviews authored by a user.
func (r *ReviewRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("WorkflowState").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

// ListAdmin returns paginated reviews for moderation.
func (r *ReviewRepository) ListAdmin(ctx context.Context, filters dto.AdminReviewListFilters) ([]models.Review, error) {
	q := r.applyAdminFilters(r.db.WithContext(ctx).Model(&models.Review{}), filters)
	var reviews []models.Review
	err := q.
		Preload("User").
		Preload("Product").
		Preload("WorkflowState").
		Order("created_at DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&reviews).Error
	return reviews, err
}
