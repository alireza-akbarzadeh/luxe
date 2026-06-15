package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
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
}

type storeService struct {
	db *gorm.DB
}

func NewStoreService(db *gorm.DB) StoreServiceInterface {
	return &storeService{db: db}
}

// ListStores returns active stores with pagination, filters and sorting.
func (s *storeService) ListStores(limit, offset int, filters dto.StoreFilter) ([]*models.Store, int64, error) {
	query := s.db.Model(&models.Store{}).Where("status = ?", "active")

	// Search (case‑insensitive)
	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Location filter
	if filters.Location != "" {
		locationTerm := "%" + strings.ToLower(filters.Location) + "%"
		query = query.Where("LOWER(location) LIKE ?", locationTerm)
	}

	// Minimum rating
	if filters.MinRating > 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}

	// Category filter (by category slug)
	if filters.CategorySlug != "" {
		query = query.Joins("JOIN store_categories sc ON sc.store_id = stores.id").
			Joins("JOIN categories c ON c.id = sc.category_id").
			Where("c.slug = ?", filters.CategorySlug)
	}

	// Count total before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	// Sorting
	switch filters.SortBy {
	case "rating":
		query = query.Order("rating DESC")
	case "followers":
		query = query.Order("follower_count DESC")
	case "newest":
		query = query.Order("joined_at DESC")
	default:
		query = query.Order("rating DESC")
	}

	// Pagination
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	// Preload categories
	var stores []*models.Store
	if err := query.Preload("Categories").Find(&stores).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return stores, total, nil
}

// GetByID retrieves a store by its primary key.
func (s *storeService) GetByID(id uint) (*models.Store, error) {
	var store models.Store
	err := s.db.Preload("Categories").First(&store, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("store not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &store, nil
}

// GetBySlug retrieves an active store by its slug.
func (s *storeService) GetBySlug(slug string) (*models.Store, error) {
	var store models.Store
	err := s.db.Where("slug = ? AND status = ?", slug, "active").
		Preload("Categories").
		First(&store).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("store not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &store, nil
}

// Create creates a new store (admin only).
func (s *storeService) Create(req dto.CreateStoreRequest) (*models.Store, error) {
	// Ensure slug is unique
	baseSlug := generateSlug(req.Name)
	slug := s.uniqSlug(baseSlug, 0)

	store := &models.Store{
		Name:          req.Name,
		Slug:          slug,
		Description:   req.Description,
		LogoURL:       req.LogoURL,
		BannerURL:     req.BannerURL,
		Location:      req.Location,
		ShippingInfo:  req.ShippingInfo,
		ReturnPolicy:  req.ReturnPolicy,
		Status:        "active",
		IsVerified:    false,
		Rating:        0,
		ReviewCount:   0,
		FollowerCount: 0,
		JoinedAt:      time.Now(),
	}

	if req.UserID != nil {
		store.UserID = req.UserID
	}

	// Associate categories if provided
	if len(req.CategoryIDs) > 0 {
		var categories []*models.Category
		if err := s.db.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		store.Categories = categories
	}

	if err := s.db.Create(store).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// Update modifies an existing store.
func (s *storeService) Update(id uint, req dto.UpdateStoreRequest) (*models.Store, error) {
	store, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		store.Name = *req.Name
		baseSlug := generateSlug(*req.Name)
		store.Slug = s.uniqSlug(baseSlug, id)
	}
	if req.Description != nil {
		store.Description = *req.Description
	}
	if req.LogoURL != nil {
		store.LogoURL = *req.LogoURL
	}
	if req.BannerURL != nil {
		store.BannerURL = *req.BannerURL
	}
	if req.Location != nil {
		store.Location = *req.Location
	}
	if req.ShippingInfo != nil {
		store.ShippingInfo = *req.ShippingInfo
	}
	if req.ReturnPolicy != nil {
		store.ReturnPolicy = *req.ReturnPolicy
	}
	if req.Status != nil {
		store.Status = *req.Status
	}
	if req.IsVerified != nil {
		store.IsVerified = *req.IsVerified
	}
	if req.CategoryIDs != nil {
		var categories []*models.Category
		if err := s.db.Where("id IN ?", req.CategoryIDs).Find(&categories).Error; err != nil {
			return nil, utils.ErrInternal(err)
		}
		store.Categories = categories
	}

	if err := s.db.Save(store).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// Delete soft-deletes a store.
func (s *storeService) Delete(id uint) error {
	result := s.db.Delete(&models.Store{}, id)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("store not found")
	}
	return nil
}

// uniqSlug ensures the slug is unique across stores.
func (s *storeService) uniqSlug(baseSlug string, excludeID uint) string {
	slug := baseSlug
	counter := 1
	for {
		var count int64
		query := s.db.Model(&models.Store{}).Where("slug = ?", slug)
		if excludeID > 0 {
			query = query.Where("id != ?", excludeID)
		}
		query.Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

// FollowStore adds a follower relationship and increments follower count.
func (s *storeService) FollowStore(userID, storeID uint) error {
	var count int64
	if err := s.db.Model(&models.StoreFollower{}).
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Count(&count).Error; err != nil {
		return utils.ErrInternal(err)
	}
	if count > 0 {
		return nil
	}

	follower := models.StoreFollower{
		UserID:  userID,
		StoreID: storeID,
	}
	if err := s.db.Create(&follower).Error; err != nil {
		return utils.ErrInternal(err)
	}

	if err := s.db.Model(&models.Store{}).
		Where("id = ?", storeID).
		Update("follower_count", gorm.Expr("follower_count + 1")).Error; err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// UnfollowStore removes a follower and decrements follower count.
func (s *storeService) UnfollowStore(userID, storeID uint) error {
	result := s.db.Where("user_id = ? AND store_id = ?", userID, storeID).
		Delete(&models.StoreFollower{})
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected > 0 {
		// Decrement follower count
		if err := s.db.Model(&models.Store{}).
			Where("id = ?", storeID).
			Update("follower_count", gorm.Expr("follower_count - 1")).Error; err != nil {
			return utils.ErrInternal(err)
		}
	}
	return nil
}

// IsFollowing checks if a user follows a store.
func (s *storeService) IsFollowing(userID, storeID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.StoreFollower{}).
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Count(&count).Error
	if err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

// GetFollowedStoreIDs returns a set of store IDs the user follows from the given list.
func (s *storeService) GetFollowedStoreIDs(userID uint, storeIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(storeIDs) == 0 {
		return result, nil
	}

	var followers []models.StoreFollower
	if err := s.db.Where("user_id = ? AND store_id IN ?", userID, storeIDs).
		Find(&followers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	for _, f := range followers {
		result[f.StoreID] = true
	}
	return result, nil
}

func (s *storeService) updateStoreReviewStats(storeID uint) {
	var result struct {
		AvgRating float64
		Count     int64
	}
	s.db.Model(&models.StoreReview{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("store_id = ?", storeID).
		Scan(&result)

	s.db.Model(&models.Store{}).Where("id = ?", storeID).
		Updates(map[string]interface{}{
			"rating":       result.AvgRating,
			"review_count": result.Count,
		})
}

func (s *storeService) buildStoreReviewSummary(storeID uint) (dto.StoreReviewSummary, error) {
	summary := dto.StoreReviewSummary{
		Counts: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
	}

	var result struct {
		AvgRating float64
		Count     int64
	}
	if err := s.db.Model(&models.StoreReview{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("store_id = ?", storeID).
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
	if err := s.db.Model(&models.StoreReview{}).
		Select("rating, COUNT(*) as count").
		Where("store_id = ?", storeID).
		Group("rating").
		Scan(&rows).Error; err != nil {
		return summary, utils.ErrInternal(err)
	}

	for _, row := range rows {
		summary.Counts[strconv.Itoa(row.Rating)] = row.Count
	}

	return summary, nil
}

// ListStoreReviews returns paginated reviews and rating summary for a store.
func (s *storeService) ListStoreReviews(storeID uint, limit, offset int) ([]models.StoreReview, int64, dto.StoreReviewSummary, error) {
	summary, err := s.buildStoreReviewSummary(storeID)
	if err != nil {
		return nil, 0, summary, err
	}

	var reviews []models.StoreReview
	query := s.db.Model(&models.StoreReview{}).Where("store_id = ?", storeID)
	if err := query.Preload("User").Order("created_at DESC").Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		return nil, 0, summary, utils.ErrInternal(err)
	}

	return reviews, summary.Total, summary, nil
}

// GetUserStoreReview returns the authenticated user's review for a store, if any.
func (s *storeService) GetUserStoreReview(userID, storeID uint) (*models.StoreReview, error) {
	var review models.StoreReview
	err := s.db.Preload("User").Where("user_id = ? AND store_id = ?", userID, storeID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return &review, nil
}

// CreateStoreReview adds a review and recalculates store rating stats.
func (s *storeService) CreateStoreReview(userID, storeID uint, req dto.CreateStoreReviewRequest) (*models.StoreReview, error) {
	var existing models.StoreReview
	err := s.db.Where("user_id = ? AND store_id = ?", userID, storeID).First(&existing).Error
	if err == nil {
		return nil, utils.ErrBadRequest("you have already reviewed this store")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	review := &models.StoreReview{
		StoreID: storeID,
		UserID:  userID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}
	if err := s.db.Create(review).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	s.updateStoreReviewStats(storeID)

	if err := s.db.Preload("User").First(review, review.ID).Error; err != nil {
		return review, nil
	}
	return review, nil
}

// UpdateStoreReview updates a review owned by the user.
func (s *storeService) UpdateStoreReview(userID, reviewID uint, req dto.UpdateStoreReviewRequest) (*models.StoreReview, error) {
	var review models.StoreReview
	err := s.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("review not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	if err := s.db.Save(&review).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	s.updateStoreReviewStats(review.StoreID)

	if err := s.db.Preload("User").First(&review, review.ID).Error; err != nil {
		return &review, nil
	}
	return &review, nil
}

// DeleteStoreReview removes a review owned by the user.
func (s *storeService) DeleteStoreReview(userID, reviewID uint) error {
	var review models.StoreReview
	err := s.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("review not found")
		}
		return utils.ErrInternal(err)
	}

	if err := s.db.Delete(&review).Error; err != nil {
		return utils.ErrInternal(err)
	}

	s.updateStoreReviewStats(review.StoreID)
	return nil
}
