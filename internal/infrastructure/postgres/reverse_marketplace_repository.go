package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// ReverseMarketplaceRepository persists buyer requests and vendor offers.
type ReverseMarketplaceRepository struct {
	db *gorm.DB
}

// NewReverseMarketplaceRepository creates a GORM-backed reverse marketplace repository.
func NewReverseMarketplaceRepository(db *gorm.DB) *ReverseMarketplaceRepository {
	return &ReverseMarketplaceRepository{db: db}
}

// ListOpen returns open buyer requests for discovery.
func (r *ReverseMarketplaceRepository) ListOpen(ctx context.Context, limit, offset int) ([]models.ReverseMarketplaceRequest, int64, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	base := r.db.WithContext(ctx).Model(&models.ReverseMarketplaceRequest{}).
		Where("status = ?", constants.ReverseMarketplaceRequestStatusOpen)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var requests []models.ReverseMarketplaceRequest
	err := base.
		Preload("Offers", func(db *gorm.DB) *gorm.DB {
			return db.Where("status <> ?", constants.ReverseMarketplaceOfferStatusWithdrawn)
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&requests).Error
	return requests, total, err
}

// ListForUser returns requests created by a buyer.
func (r *ReverseMarketplaceRepository) ListForUser(ctx context.Context, userID uint, limit, offset int) ([]models.ReverseMarketplaceRequest, int64, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	base := r.db.WithContext(ctx).Model(&models.ReverseMarketplaceRequest{}).
		Where("user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var requests []models.ReverseMarketplaceRequest
	err := base.
		Preload("Offers", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Offers.Store").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&requests).Error
	return requests, total, err
}

// GetByID loads a request with offers and store names.
func (r *ReverseMarketplaceRepository) GetByID(ctx context.Context, id uint) (*models.ReverseMarketplaceRequest, error) {
	var request models.ReverseMarketplaceRequest
	err := r.db.WithContext(ctx).
		Preload("Offers", func(db *gorm.DB) *gorm.DB {
			return db.Where("status <> ?", constants.ReverseMarketplaceOfferStatusWithdrawn).
				Order("created_at ASC")
		}).
		Preload("Offers.Store").
		First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// Create inserts a buyer request.
func (r *ReverseMarketplaceRepository) Create(ctx context.Context, request *models.ReverseMarketplaceRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

// CreateOffer inserts a vendor offer for a request.
func (r *ReverseMarketplaceRepository) CreateOffer(ctx context.Context, offer *models.ReverseMarketplaceOffer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

// GetOfferByStoreAndRequest finds an existing offer for idempotency checks.
func (r *ReverseMarketplaceRepository) GetOfferByStoreAndRequest(ctx context.Context, storeID, requestID uint) (*models.ReverseMarketplaceOffer, error) {
	var offer models.ReverseMarketplaceOffer
	err := r.db.WithContext(ctx).
		Where("store_id = ? AND request_id = ?", storeID, requestID).
		First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}
