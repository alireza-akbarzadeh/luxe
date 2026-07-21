package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HomeRepository persists homepage personalization and merchandising config.
type HomeRepository struct {
	db *gorm.DB
}

// NewHomeRepository creates a GORM-backed home repository.
func NewHomeRepository(db *gorm.DB) *HomeRepository {
	return &HomeRepository{db: db}
}

// ListFavoriteCategoryIDs returns category ids saved by the user.
func (r *HomeRepository) ListFavoriteCategoryIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&models.UserFavoriteCategory{}).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Pluck("category_id", &ids).Error
	return ids, err
}

// ReplaceFavoriteCategories replaces all favorite categories for a user.
func (r *HomeRepository) ReplaceFavoriteCategories(ctx context.Context, userID uint, categoryIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserFavoriteCategory{}).Error; err != nil {
			return err
		}
		if len(categoryIDs) == 0 {
			return nil
		}
		rows := make([]models.UserFavoriteCategory, 0, len(categoryIDs))
		for _, cid := range categoryIDs {
			rows = append(rows, models.UserFavoriteCategory{UserID: userID, CategoryID: cid})
		}
		return tx.Create(&rows).Error
	})
}

// ListFavoriteCategories loads favorite categories with metadata.
func (r *HomeRepository) ListFavoriteCategories(ctx context.Context, userID uint) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN user_favorite_categories ufc ON ufc.category_id = categories.id AND ufc.user_id = ?", userID).
		Where("categories.deleted_at IS NULL AND categories.is_active = ?", true).
		Order("ufc.created_at ASC").
		Find(&categories).Error
	return categories, err
}

// UpsertProductView records or refreshes a product view timestamp.
func (r *HomeRepository) UpsertProductView(ctx context.Context, userID, productID uint) error {
	row := models.UserProductView{
		UserID:    userID,
		ProductID: productID,
		ViewedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&row).Error
}

// ListRecentlyViewedProductIDs returns recently viewed product ids for a user.
func (r *HomeRepository) ListRecentlyViewedProductIDs(ctx context.Context, userID uint, limit int) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&models.UserProductView{}).
		Where("user_id = ?", userID).
		Order("viewed_at DESC").
		Limit(limit).
		Pluck("product_id", &ids).Error
	return ids, err
}

// ListActiveFlashDeals returns active flash deals with products.
func (r *HomeRepository) ListActiveFlashDeals(ctx context.Context, limit int) ([]models.FlashDeal, error) {
	now := time.Now()
	var deals []models.FlashDeal
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Brand").
		Where("status = ? AND ends_at > ? AND (starts_at IS NULL OR starts_at <= ?)", "active", now, now).
		Order("sort_order ASC, ends_at ASC").
		Limit(limit).
		Find(&deals).Error
	return deals, err
}

// ListPublishedHomepageSections returns published homepage section config rows.
func (r *HomeRepository) ListPublishedHomepageSections(ctx context.Context, limit int) ([]models.HomepageSection, error) {
	return r.listPublishedHomepageSections(ctx, limit, false)
}

// ListPublishedSeasonalSections returns published sections excluding hero slides and flash promo config.
func (r *HomeRepository) ListPublishedSeasonalSections(ctx context.Context, limit int) ([]models.HomepageSection, error) {
	return r.listPublishedHomepageSections(ctx, limit, true)
}

func (r *HomeRepository) listPublishedHomepageSections(ctx context.Context, limit int, excludeHeroAndPromo bool) ([]models.HomepageSection, error) {
	var rows []models.HomepageSection
	query := r.db.WithContext(ctx).Where("status = ?", "published")
	if excludeHeroAndPromo {
		query = query.Where(
			"section_key NOT LIKE ? AND section_key NOT LIKE ? AND section_key <> ?",
			"hero-%", "marketing-band-%", "flash-deals-promo",
		)
	}
	err := query.Order("sort_order ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

// ListPublishedHeroSlides returns homepage sections for the hero carousel.
func (r *HomeRepository) ListPublishedHeroSlides(ctx context.Context, limit int) ([]models.HomepageSection, error) {
	var rows []models.HomepageSection
	err := r.db.WithContext(ctx).
		Where("status = ? AND section_key LIKE ?", "published", "hero-%").
		Order("sort_order ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// FindPublishedHomepageSectionByKey returns one published section by exact key.
func (r *HomeRepository) FindPublishedHomepageSectionByKey(ctx context.Context, key string) (*models.HomepageSection, error) {
	var row models.HomepageSection
	err := r.db.WithContext(ctx).
		Where("status = ? AND section_key = ?", "published", key).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListPublishedMarketingBands returns promo band config rows for the storefront home page.
func (r *HomeRepository) ListPublishedMarketingBands(ctx context.Context, limit int) ([]models.HomepageSection, error) {
	var rows []models.HomepageSection
	err := r.db.WithContext(ctx).
		Where("status = ?", "published").
		Where("section_key = ? OR section_key LIKE ?", "flash-deals-promo", "marketing-band-%").
		Order("sort_order ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// ListPublishedCollections returns active collections within their schedule window for homepage.
func (r *HomeRepository) ListPublishedCollections(ctx context.Context, limit int) ([]models.Collection, error) {
	var rows []models.Collection
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Where("(starts_at IS NULL OR starts_at <= ?)", now).
		Where("(ends_at IS NULL OR ends_at >= ?)", now).
		Order("sort_order ASC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// ListUserOrderCategoryIDs returns category ids from user's past orders for recommendations.
func (r *HomeRepository) ListUserOrderCategoryIDs(ctx context.Context, userID uint, limit int) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("order_items oi").
		Select("DISTINCT p.category_id").
		Joins("INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products p ON p.id = oi.product_id AND p.deleted_at IS NULL").
		Where("o.user_id = ? AND p.category_id IS NOT NULL", userID).
		Order("o.created_at DESC").
		Limit(limit).
		Pluck("p.category_id", &ids).Error
	return ids, err
}

// ListUserLikedCategoryIDs returns distinct category ids from a user's wishlisted products.
func (r *HomeRepository) ListUserLikedCategoryIDs(ctx context.Context, userID uint, limit int) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("product_likes pl").
		Select("DISTINCT p.category_id").
		Joins("INNER JOIN products p ON p.id = pl.product_id AND p.deleted_at IS NULL").
		Where("pl.user_id = ? AND p.category_id IS NOT NULL", userID).
		Order("pl.created_at DESC").
		Limit(limit).
		Pluck("p.category_id", &ids).Error
	return ids, err
}

// ListUserLikedProductIDs returns wishlisted product ids for recommendations.
func (r *HomeRepository) ListUserLikedProductIDs(ctx context.Context, userID uint, limit int) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&models.ProductLike{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Pluck("product_id", &ids).Error
	return ids, err
}
