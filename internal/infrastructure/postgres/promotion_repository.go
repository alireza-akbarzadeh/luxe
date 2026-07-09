package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PromotionRepository handles admin CRUD for merchandising promotions.
type PromotionRepository struct {
	db *gorm.DB
}

// NewPromotionRepository creates a promotion repository.
func NewPromotionRepository(db *gorm.DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

// --- Flash deals ---

func (r *PromotionRepository) ListFlashDeals(ctx context.Context, filters dto.AdminFlashDealListFilters) ([]models.FlashDeal, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.FlashDeal{})
	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("title ILIKE ? OR CAST(product_id AS TEXT) = ?", like, strings.TrimSpace(filters.Search))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.FlashDeal
	err := query.Preload("Product").
		Order("sort_order ASC, ends_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	return rows, total, err
}

func (r *PromotionRepository) FindFlashDealByID(ctx context.Context, id uint) (*models.FlashDeal, error) {
	var row models.FlashDeal
	err := r.db.WithContext(ctx).Preload("Product").First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PromotionRepository) CreateFlashDeal(ctx context.Context, row *models.FlashDeal) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PromotionRepository) UpdateFlashDeal(ctx context.Context, row *models.FlashDeal) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *PromotionRepository) DeleteFlashDeal(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.FlashDeal{}, id).Error
}

// --- Homepage sections (featured banners) ---

func (r *PromotionRepository) ListHomepageSections(ctx context.Context, filters dto.AdminHomepageSectionListFilters) ([]models.HomepageSection, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.HomepageSection{})
	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("title ILIKE ? OR section_key ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.HomepageSection
	err := query.Order("sort_order ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	return rows, total, err
}

func (r *PromotionRepository) FindHomepageSectionByID(ctx context.Context, id uint) (*models.HomepageSection, error) {
	var row models.HomepageSection
	err := r.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PromotionRepository) FindHomepageSectionByKey(ctx context.Context, key string, excludeID uint) (bool, error) {
	query := r.db.WithContext(ctx).Model(&models.HomepageSection{}).Where("section_key = ?", key)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *PromotionRepository) CreateHomepageSection(ctx context.Context, row *models.HomepageSection) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PromotionRepository) UpdateHomepageSection(ctx context.Context, row *models.HomepageSection) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *PromotionRepository) DeleteHomepageSection(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.HomepageSection{}, id).Error
}

// --- Campaigns ---

func (r *PromotionRepository) ListCampaigns(ctx context.Context, filters dto.AdminCampaignListFilters) ([]models.Campaign, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.Campaign{})
	if filters.Status != "" && filters.Status != "all" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Search != "" {
		like := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.Campaign
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	return rows, total, err
}

func (r *PromotionRepository) FindCampaignByID(ctx context.Context, id uint) (*models.Campaign, error) {
	var row models.Campaign
	err := r.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PromotionRepository) FindCampaignBySlug(ctx context.Context, slug string, excludeID uint) (bool, error) {
	query := r.db.WithContext(ctx).Model(&models.Campaign{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *PromotionRepository) CreateCampaign(ctx context.Context, row *models.Campaign) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *PromotionRepository) UpdateCampaign(ctx context.Context, row *models.Campaign) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *PromotionRepository) DeleteCampaign(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Campaign{}, id).Error
}

// CountActivePromotions returns KPI counts for the admin promotions hub.
func (r *PromotionRepository) CountActivePromotions(ctx context.Context, now time.Time) (flashActive int64, bannersPublished int64, campaignsActive int64, err error) {
	if err = r.db.WithContext(ctx).Model(&models.FlashDeal{}).
		Where("status = ? AND ends_at > ? AND (starts_at IS NULL OR starts_at <= ?)", "active", now, now).
		Count(&flashActive).Error; err != nil {
		return
	}
	if err = r.db.WithContext(ctx).Model(&models.HomepageSection{}).
		Where("status = ?", "published").
		Count(&bannersPublished).Error; err != nil {
		return
	}
	err = r.db.WithContext(ctx).Model(&models.Campaign{}).
		Where("status = ?", "active").
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now).
		Count(&campaignsActive).Error
	return
}
