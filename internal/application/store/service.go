package store

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates store read use cases.
type Queries struct {
	repo *postgres.StoreRepository
}

// NewQueries creates store query use cases.
func NewQueries(repo *postgres.StoreRepository) *Queries {
	return &Queries{repo: repo}
}

// ListStores returns active stores with pagination, filters and sorting.
func (q *Queries) ListStores(limit, offset int, filters dto.StoreFilter) ([]*models.Store, int64, error) {
	query := q.repo.ListActive(filters)

	total, err := q.repo.CountQuery(query)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	query = postgres.ApplyStoreSort(query, filters.SortBy)
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	stores, err := q.repo.FindStores(query)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return stores, total, nil
}

// ListVendorStores returns stores the current user can manage in the vendor panel.
func (q *Queries) ListVendorStores(ctx context.Context, userID uint, role string) ([]*models.Store, error) {
	query := q.repo.ListVendorStoresQuery(ctx, userID, role)
	stores, err := q.repo.FindStores(query)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return stores, nil
}

// GetByID retrieves a store by its primary key.
func (q *Queries) GetByID(id uint) (*models.Store, error) {
	store, err := q.repo.FindByID(id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("store not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// GetBySlug retrieves an active store by its slug.
func (q *Queries) GetBySlug(slug string) (*models.Store, error) {
	store, err := q.repo.FindActiveBySlug(slug)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("store not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// IsFollowing checks if a user follows a store.
func (q *Queries) IsFollowing(userID, storeID uint) (bool, error) {
	count, err := q.repo.CountFollower(userID, storeID)
	if err != nil {
		return false, utils.ErrInternal(err)
	}
	return count > 0, nil
}

// GetFollowedStoreIDs returns a set of store IDs the user follows from the given list.
func (q *Queries) GetFollowedStoreIDs(userID uint, storeIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if len(storeIDs) == 0 {
		return result, nil
	}

	followers, err := q.repo.FindFollowersByUserAndStoreIDs(userID, storeIDs)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	for _, f := range followers {
		result[f.StoreID] = true
	}
	return result, nil
}

// ListStoreReviews returns paginated reviews and rating summary for a store.
func (q *Queries) ListStoreReviews(storeID uint, limit, offset int) ([]models.StoreReview, int64, dto.StoreReviewSummary, error) {
	summary, err := q.buildStoreReviewSummary(storeID)
	if err != nil {
		return nil, 0, summary, err
	}

	reviews, err := q.repo.FindStoreReviews(storeID, limit, offset)
	if err != nil {
		return nil, 0, summary, utils.ErrInternal(err)
	}
	return reviews, summary.Total, summary, nil
}

// GetUserStoreReview returns the authenticated user's review for a store, if any.
func (q *Queries) GetUserStoreReview(userID, storeID uint) (*models.StoreReview, error) {
	review, err := q.repo.FindUserStoreReview(userID, storeID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, nil
		}
		return nil, utils.ErrInternal(err)
	}
	return review, nil
}

func (q *Queries) buildStoreReviewSummary(storeID uint) (dto.StoreReviewSummary, error) {
	summary := dto.StoreReviewSummary{
		Counts: map[string]int{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0},
	}

	avg, count, err := q.repo.ScanReviewStats(storeID)
	if err != nil {
		return summary, utils.ErrInternal(err)
	}
	summary.Average = avg
	summary.Total = count

	counts, err := q.repo.ScanReviewRatingCounts(storeID)
	if err != nil {
		return summary, utils.ErrInternal(err)
	}
	for k, v := range counts {
		summary.Counts[k] = v
	}
	return summary, nil
}

// Commands orchestrates store write use cases.
type Commands struct {
	repo    *postgres.StoreRepository
	queries *Queries
}

// NewCommands creates store command use cases.
func NewCommands(repo *postgres.StoreRepository, queries *Queries) *Commands {
	return &Commands{repo: repo, queries: queries}
}

// Create creates a new store.
func (c *Commands) Create(req dto.CreateStoreRequest) (*models.Store, error) {
	baseSlug := postgres.GenerateSlug(req.Name)
	slug, err := c.repo.UniqueSlug(baseSlug, 0)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	store := &models.Store{
		Name:          req.Name,
		Slug:          slug,
		Description:   req.Description,
		LogoURL:       req.LogoURL,
		BannerURL:     req.BannerURL,
		Location:      req.Location,
		ShippingInfo:  req.ShippingInfo,
		ReturnPolicy:  req.ReturnPolicy,
		Status:        constants.StoreStatusActive,
		IsVerified:    false,
		Rating:        0,
		ReviewCount:   0,
		FollowerCount: 0,
		JoinedAt:      postgres.DefaultJoinedAt(),
	}

	if req.UserID != nil {
		store.UserID = req.UserID
	}

	if len(req.CategoryIDs) > 0 {
		categories, err := c.repo.FindCategoriesByIDs(req.CategoryIDs)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		store.Categories = categories
	}

	if err := c.repo.Create(store); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// Update modifies an existing store.
func (c *Commands) Update(id uint, req dto.UpdateStoreRequest) (*models.Store, error) {
	store, err := c.queries.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		store.Name = *req.Name
		baseSlug := postgres.GenerateSlug(*req.Name)
		slug, slugErr := c.repo.UniqueSlug(baseSlug, id)
		if slugErr != nil {
			return nil, utils.ErrInternal(slugErr)
		}
		store.Slug = slug
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
		categories, catErr := c.repo.FindCategoriesByIDs(*req.CategoryIDs)
		if catErr != nil {
			return nil, utils.ErrInternal(catErr)
		}
		store.Categories = categories
	}

	if err := c.repo.Save(store); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return store, nil
}

// Delete soft-deletes a store.
func (c *Commands) Delete(id uint) error {
	rows, err := c.repo.DeleteByID(id)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("store not found")
	}
	return nil
}

// FollowStore adds a follower relationship and increments follower count.
func (c *Commands) FollowStore(userID, storeID uint) error {
	count, err := c.repo.CountFollower(userID, storeID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if count > 0 {
		return nil
	}

	follower := models.StoreFollower{
		UserID:  userID,
		StoreID: storeID,
	}
	if err := c.repo.CreateFollower(&follower); err != nil {
		return utils.ErrInternal(err)
	}
	return c.repo.IncrementFollowerCount(storeID)
}

// UnfollowStore removes a follower and decrements follower count.
func (c *Commands) UnfollowStore(userID, storeID uint) error {
	rows, err := c.repo.DeleteFollower(userID, storeID)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows > 0 {
		return c.repo.DecrementFollowerCount(storeID)
	}
	return nil
}

// CreateStoreReview adds a review and recalculates store rating stats.
func (c *Commands) CreateStoreReview(userID, storeID uint, req dto.CreateStoreReviewRequest) (*models.StoreReview, error) {
	_, err := c.repo.FindExistingStoreReview(userID, storeID)
	if err == nil {
		return nil, utils.ErrBadRequest("you have already reviewed this store")
	}
	if !postgres.IsNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	review := &models.StoreReview{
		StoreID: storeID,
		UserID:  userID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}
	if err := c.repo.CreateStoreReview(review); err != nil {
		return nil, utils.ErrInternal(err)
	}

	c.updateStoreReviewStats(storeID)

	if err := c.repo.PreloadUserOnReview(review); err != nil {
		return review, nil
	}
	return review, nil
}

// UpdateStoreReview updates a review owned by the user.
func (c *Commands) UpdateStoreReview(userID, reviewID uint, req dto.UpdateStoreReviewRequest) (*models.StoreReview, error) {
	review, err := c.repo.FindStoreReviewByUser(userID, reviewID)
	if err != nil {
		if postgres.IsNotFound(err) {
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

	if err := c.repo.SaveStoreReview(review); err != nil {
		return nil, utils.ErrInternal(err)
	}

	c.updateStoreReviewStats(review.StoreID)

	if err := c.repo.PreloadUserOnReview(review); err != nil {
		return review, nil
	}
	return review, nil
}

// DeleteStoreReview removes a review owned by the user.
func (c *Commands) DeleteStoreReview(userID, reviewID uint) error {
	review, err := c.repo.FindStoreReviewByUser(userID, reviewID)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("review not found")
		}
		return utils.ErrInternal(err)
	}

	if err := c.repo.DeleteStoreReview(review); err != nil {
		return utils.ErrInternal(err)
	}

	c.updateStoreReviewStats(review.StoreID)
	return nil
}

func (c *Commands) updateStoreReviewStats(storeID uint) {
	avg, count, err := c.repo.ScanReviewStats(storeID)
	if err != nil {
		return
	}
	_ = c.repo.UpdateStoreReviewStats(storeID, avg, count)
}
