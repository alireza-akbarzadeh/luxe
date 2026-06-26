package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// StoreRepository persists stores and related entities with GORM.
type StoreRepository struct {
	db *gorm.DB
}

// NewStoreRepository creates a GORM-backed store repository.
func NewStoreRepository(db *gorm.DB) *StoreRepository {
	return &StoreRepository{db: db}
}

// ListActive builds a filtered query for active stores.
func (r *StoreRepository) ListActive(filters dto.StoreFilter) *gorm.DB {
	query := r.db.Model(&models.Store{}).Where("status = ?", constants.StoreStatusActive)

	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}
	if filters.Location != "" {
		locationTerm := "%" + strings.ToLower(filters.Location) + "%"
		query = query.Where("LOWER(location) LIKE ?", locationTerm)
	}
	if filters.MinRating > 0 {
		query = query.Where("rating >= ?", filters.MinRating)
	}
	if filters.CategorySlug != "" {
		query = query.Joins("JOIN store_categories sc ON sc.store_id = stores.id").
			Joins("JOIN categories c ON c.id = sc.category_id").
			Where("c.slug = ?", filters.CategorySlug)
	}
	return query
}

// CountQuery counts rows for a scoped query.
func (r *StoreRepository) CountQuery(query *gorm.DB) (int64, error) {
	var total int64
	err := query.Count(&total).Error
	return total, err
}

// FindStores executes a store list query with preloaded categories.
func (r *StoreRepository) FindStores(query *gorm.DB) ([]*models.Store, error) {
	var stores []*models.Store
	err := query.Preload("Categories").Find(&stores).Error
	return stores, err
}

// ListVendorStoresQuery builds vendor store list query scoped by role.
func (r *StoreRepository) ListVendorStoresQuery(ctx context.Context, userID uint, role string) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&models.Store{}).Preload("Categories").Order("name ASC")
	if role != constants.RoleAdmin && role != constants.RoleModerator {
		query = query.Where("user_id = ?", userID)
	}
	return query
}

// FindByID loads a store with categories preloaded.
func (r *StoreRepository) FindByID(id uint) (*models.Store, error) {
	var store models.Store
	err := r.db.Preload("Categories").Preload("User").First(&store, id).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// FindActiveBySlug loads an active store by slug.
func (r *StoreRepository) FindActiveBySlug(slug string) (*models.Store, error) {
	var store models.Store
	err := r.db.Where("slug = ? AND status = ?", slug, constants.StoreStatusActive).
		Preload("Categories").
		First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// FindCategoriesByIDs loads categories by id list.
func (r *StoreRepository) FindCategoriesByIDs(ids []uint) ([]*models.Category, error) {
	var categories []*models.Category
	err := r.db.Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}

// Create inserts a store row.
func (r *StoreRepository) Create(store *models.Store) error {
	return r.db.Create(store).Error
}

// Save persists store changes.
func (r *StoreRepository) Save(store *models.Store) error {
	return r.db.Save(store).Error
}

// DeleteByID soft-deletes a store.
func (r *StoreRepository) DeleteByID(id uint) (int64, error) {
	result := r.db.Delete(&models.Store{}, id)
	return result.RowsAffected, result.Error
}

// ListAdmin builds a filtered admin store list query.
func (r *StoreRepository) ListAdmin(filters dto.AdminStoreFilter) *gorm.DB {
	query := r.db.Model(&models.Store{}).Preload("User").Preload("Categories")

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}
	if filters.IsVerified != nil {
		query = query.Where("is_verified = ?", *filters.IsVerified)
	}

	switch filters.SortBy {
	case "oldest":
		query = query.Order("created_at ASC")
	default:
		query = query.Order("created_at DESC")
	}

	return query
}

// FindAdminStores executes an admin store list query.
func (r *StoreRepository) FindAdminStores(query *gorm.DB) ([]*models.Store, error) {
	var stores []*models.Store
	err := query.Find(&stores).Error
	return stores, err
}

// FindOwnedStore loads a store scoped to the owner unless admin access is allowed.
func (r *StoreRepository) FindOwnedStore(ctx context.Context, storeID, userID uint, allowAdmin bool) (*models.Store, error) {
	query := r.db.WithContext(ctx).Preload("Categories").Preload("User")
	if !allowAdmin {
		query = query.Where("user_id = ?", userID)
	}

	var store models.Store
	err := query.First(&store, storeID).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

// CountBySlug counts stores with a slug, optionally excluding an id.
func (r *StoreRepository) CountBySlug(slug string, excludeID uint) (int64, error) {
	var count int64
	query := r.db.Model(&models.Store{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

// CountFollower checks if a user follows a store.
func (r *StoreRepository) CountFollower(userID, storeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.StoreFollower{}).
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Count(&count).Error
	return count, err
}

// CreateFollower inserts a store follower row.
func (r *StoreRepository) CreateFollower(follower *models.StoreFollower) error {
	return r.db.Create(follower).Error
}

// DeleteFollower removes a store follower row.
func (r *StoreRepository) DeleteFollower(userID, storeID uint) (int64, error) {
	result := r.db.Where("user_id = ? AND store_id = ?", userID, storeID).
		Delete(&models.StoreFollower{})
	return result.RowsAffected, result.Error
}

// IncrementFollowerCount increments follower_count for a store.
func (r *StoreRepository) IncrementFollowerCount(storeID uint) error {
	return r.db.Model(&models.Store{}).
		Where("id = ?", storeID).
		Update("follower_count", gorm.Expr("follower_count + 1")).Error
}

// DecrementFollowerCount decrements follower_count for a store.
func (r *StoreRepository) DecrementFollowerCount(storeID uint) error {
	return r.db.Model(&models.Store{}).
		Where("id = ?", storeID).
		Update("follower_count", gorm.Expr("follower_count - 1")).Error
}

// FindFollowersByUserAndStoreIDs loads follower rows for given store ids.
func (r *StoreRepository) FindFollowersByUserAndStoreIDs(userID uint, storeIDs []uint) ([]models.StoreFollower, error) {
	var followers []models.StoreFollower
	err := r.db.Where("user_id = ? AND store_id IN ?", userID, storeIDs).
		Find(&followers).Error
	return followers, err
}

// ScanReviewStats aggregates average rating and count for a store.
func (r *StoreRepository) ScanReviewStats(storeID uint) (avgRating float64, count int64, err error) {
	var result struct {
		AvgRating float64
		Count     int64
	}
	err = r.db.Model(&models.StoreReview{}).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as count").
		Where("store_id = ?", storeID).
		Scan(&result).Error
	return result.AvgRating, result.Count, err
}

// UpdateStoreReviewStats writes rating aggregates to the store row.
func (r *StoreRepository) UpdateStoreReviewStats(storeID uint, avgRating float64, count int64) error {
	return r.db.Model(&models.Store{}).Where("id = ?", storeID).
		Updates(map[string]interface{}{
			"rating":       avgRating,
			"review_count": count,
		}).Error
}

// ScanReviewRatingCounts returns per-rating counts for a store.
func (r *StoreRepository) ScanReviewRatingCounts(storeID uint) (map[string]int, error) {
	type ratingCount struct {
		Rating int
		Count  int
	}
	var rows []ratingCount
	err := r.db.Model(&models.StoreReview{}).
		Select("rating, COUNT(*) as count").
		Where("store_id = ?", storeID).
		Group("rating").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}
	for _, row := range rows {
		counts[strconv.Itoa(row.Rating)] = row.Count
	}
	return counts, nil
}

// FindStoreReviews returns paginated store reviews with user preloaded.
func (r *StoreRepository) FindStoreReviews(storeID uint, limit, offset int) ([]models.StoreReview, error) {
	var reviews []models.StoreReview
	err := r.db.Model(&models.StoreReview{}).
		Where("store_id = ?", storeID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

// FindUserStoreReview loads a user's review for a store.
func (r *StoreRepository) FindUserStoreReview(userID, storeID uint) (*models.StoreReview, error) {
	var review models.StoreReview
	err := r.db.Preload("User").Where("user_id = ? AND store_id = ?", userID, storeID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindStoreReviewByUser loads a review owned by the user.
func (r *StoreRepository) FindStoreReviewByUser(userID, reviewID uint) (*models.StoreReview, error) {
	var review models.StoreReview
	err := r.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// FindExistingStoreReview checks for an existing review by user and store.
func (r *StoreRepository) FindExistingStoreReview(userID, storeID uint) (*models.StoreReview, error) {
	var existing models.StoreReview
	err := r.db.Where("user_id = ? AND store_id = ?", userID, storeID).First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

// CreateStoreReview inserts a store review.
func (r *StoreRepository) CreateStoreReview(review *models.StoreReview) error {
	return r.db.Create(review).Error
}

// SaveStoreReview persists review changes.
func (r *StoreRepository) SaveStoreReview(review *models.StoreReview) error {
	return r.db.Save(review).Error
}

// DeleteStoreReview removes a review row.
func (r *StoreRepository) DeleteStoreReview(review *models.StoreReview) error {
	return r.db.Delete(review).Error
}

// PreloadUserOnReview reloads a review with user preloaded.
func (r *StoreRepository) PreloadUserOnReview(review *models.StoreReview) error {
	return r.db.Preload("User").First(review, review.ID).Error
}

// GenerateSlug normalizes a name into a URL slug.
func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

// UniqueSlug finds an unused slug, optionally excluding a store id.
func (r *StoreRepository) UniqueSlug(baseSlug string, excludeID uint) (string, error) {
	slug := baseSlug
	counter := 1
	for {
		count, err := r.CountBySlug(slug, excludeID)
		if err != nil {
			return "", err
		}
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug, nil
}

// ApplyStoreSort applies sort order to a store list query.
func ApplyStoreSort(query *gorm.DB, sortBy string) *gorm.DB {
	switch sortBy {
	case "rating":
		return query.Order("rating DESC")
	case "followers":
		return query.Order("follower_count DESC")
	case "newest":
		return query.Order("joined_at DESC")
	default:
		return query.Order("rating DESC")
	}
}

// DefaultJoinedAt returns the current time for new stores.
func DefaultJoinedAt() time.Time {
	return time.Now()
}
