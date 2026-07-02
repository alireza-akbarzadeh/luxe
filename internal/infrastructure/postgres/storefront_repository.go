package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// StorefrontRepository serves curated landing-page catalog queries.
type StorefrontRepository struct {
	db *gorm.DB
}

// NewStorefrontRepository creates a GORM-backed storefront repository.
func NewStorefrontRepository(db *gorm.DB) *StorefrontRepository {
	return &StorefrontRepository{db: db}
}

var storefrontProductPreloads = []string{"Category", "Brand", "Attributes", "WorkflowState"}

func (r *StorefrontRepository) preloadProducts(q *gorm.DB) *gorm.DB {
	for _, name := range storefrontProductPreloads {
		q = q.Preload(name)
	}
	return q
}

// ListDiscountedProducts returns active products on sale ordered by discount depth.
func (r *StorefrontRepository) ListDiscountedProducts(ctx context.Context, limit, offset int) ([]*models.Product, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("status = ?", constants.ProductStatusActive).
		Where("compare_at_price IS NOT NULL AND compare_at_price > price")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []*models.Product
	q := r.preloadProducts(base).
		Order("(compare_at_price - price) / NULLIF(compare_at_price, 0) DESC").
		Order("reviews_count DESC").
		Limit(limit).Offset(offset)
	if err := q.Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

type bestSellerRow struct {
	ProductID uint
	UnitsSold int64
}

// ListBestSellerIDs returns product ids ranked by units sold in paid orders.
func (r *StorefrontRepository) ListBestSellerIDs(ctx context.Context, limit int) ([]bestSellerRow, error) {
	since := time.Now().AddDate(0, -6, 0)
	var rows []bestSellerRow
	err := r.db.WithContext(ctx).Model(&models.OrderItem{}).
		Select(`
			order_items.product_id AS product_id,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold`).
		Joins("INNER JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL").
		Joins("INNER JOIN products p ON p.id = order_items.product_id AND p.deleted_at IS NULL").
		Where("o.created_at >= ? AND o.status IN ?", since, revenueOrderStatuses).
		Where("p.status = ?", constants.ProductStatusActive).
		Group("order_items.product_id").
		Order("units_sold DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ListProductsByIDs loads products preserving the given id order.
func (r *StorefrontRepository) ListProductsByIDs(ctx context.Context, ids []uint) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}
	var products []*models.Product
	q := r.preloadProducts(r.db.WithContext(ctx))
	if err := q.Where("id IN ?", ids).Find(&products).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]*models.Product, len(products))
	for _, p := range products {
		byID[p.ID] = p
	}
	ordered := make([]*models.Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered, nil
}

type popularBrandRow struct {
	models.Brand
	ProductCount int64 `gorm:"column:product_count"`
}

// ListPopularBrands returns brands with the most active catalog products.
func (r *StorefrontRepository) ListPopularBrands(ctx context.Context, limit int) ([]popularBrandRow, error) {
	var rows []popularBrandRow
	err := r.db.WithContext(ctx).Table("brands").
		Select("brands.*, COUNT(products.id) AS product_count").
		Joins("INNER JOIN products ON products.brand_id = brands.id AND products.deleted_at IS NULL").
		Where("products.status = ?", constants.ProductStatusActive).
		Where("brands.status IN ?", []string{"active", "published"}).
		Group("brands.id").
		Order("product_count DESC").
		Order("brands.name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

type homeCategoryRow struct {
	models.Category
	ProductCount int64 `gorm:"column:product_count"`
}

// ListHomeCategories returns active top-level categories ranked by catalog size.
func (r *StorefrontRepository) ListHomeCategories(ctx context.Context, limit int) ([]homeCategoryRow, error) {
	var rows []homeCategoryRow
	err := r.db.WithContext(ctx).Table("categories").
		Select("categories.*, COUNT(products.id) AS product_count").
		Joins("LEFT JOIN products ON products.category_id = categories.id AND products.deleted_at IS NULL AND products.status = ?", constants.ProductStatusActive).
		Where("categories.is_active = ?", true).
		Where("categories.deleted_at IS NULL").
		Where("categories.parent_id IS NULL").
		Group("categories.id").
		Order("product_count DESC").
		Order("categories.name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// FindCategoryCoverImages returns the first product image per category id.
func (r *StorefrontRepository) FindCategoryCoverImages(ctx context.Context, categoryIDs []uint) (map[uint]string, error) {
	out := make(map[uint]string, len(categoryIDs))
	if len(categoryIDs) == 0 {
		return out, nil
	}

	var products []models.Product
	err := r.db.WithContext(ctx).
		Select("category_id", "images").
		Where("category_id IN ?", categoryIDs).
		Where("status = ?", constants.ProductStatusActive).
		Where("array_length(images, 1) > 0").
		Order("rating DESC, reviews_count DESC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.CategoryID == nil {
			continue
		}
		if _, exists := out[*p.CategoryID]; exists {
			continue
		}
		if len(p.Images) > 0 && p.Images[0] != "" {
			out[*p.CategoryID] = p.Images[0]
		}
	}
	return out, nil
}

type topBrandRow struct {
	models.Brand
	ProductCount int64   `gorm:"column:product_count"`
	UnitsSold    int64   `gorm:"column:units_sold"`
	Revenue      float64 `gorm:"column:revenue"`
	MinPrice     float64 `gorm:"column:min_price"`
}

// ListTopBrandsBySales ranks brands by order revenue and units sold.
func (r *StorefrontRepository) ListTopBrandsBySales(ctx context.Context, limit int) ([]topBrandRow, error) {
	since := time.Now().AddDate(0, -6, 0)
	var rows []topBrandRow
	err := r.db.WithContext(ctx).Table("brands").
		Select(`
			brands.*,
			COUNT(DISTINCT products.id) AS product_count,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue,
			COALESCE(MIN(products.price), 0) AS min_price`).
		Joins("INNER JOIN products ON products.brand_id = brands.id AND products.deleted_at IS NULL AND products.status = ?", constants.ProductStatusActive).
		Joins("LEFT JOIN order_items ON order_items.product_id = products.id").
		Joins("LEFT JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL AND o.created_at >= ? AND o.status IN ?", since, revenueOrderStatuses).
		Where("brands.status IN ?", []string{"active", "published"}).
		Group("brands.id").
		Order("revenue DESC, units_sold DESC, product_count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

type trendingRow struct {
	ProductID uint
	Score     float64
}

// ListTrendingProductIDs scores products by recent sales and wishlist activity.
func (r *StorefrontRepository) ListTrendingProductIDs(ctx context.Context, limit int) ([]trendingRow, error) {
	since := time.Now().AddDate(0, 0, -14)
	var rows []trendingRow
	err := r.db.WithContext(ctx).Table("products p").
		Select(`
			p.id AS product_id,
			(COALESCE(sales.units_sold, 0) * 3 + COALESCE(likes.like_count, 0) * 2 + COALESCE(views.view_count, 0)) AS score`).
		Joins(`LEFT JOIN (
			SELECT oi.product_id, SUM(oi.quantity) AS units_sold
			FROM order_items oi
			INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL
			WHERE o.created_at >= ? AND o.status IN ?
			GROUP BY oi.product_id
		) sales ON sales.product_id = p.id`, since, revenueOrderStatuses).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) AS like_count
			FROM product_likes
			GROUP BY product_id
		) likes ON likes.product_id = p.id`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) AS view_count
			FROM user_product_views
			WHERE viewed_at >= ?
			GROUP BY product_id
		) views ON views.product_id = p.id`, since).
		Where("p.deleted_at IS NULL AND p.status = ?", constants.ProductStatusActive).
		Order("score DESC, p.reviews_count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ListNewArrivalProducts returns newest active products.
func (r *StorefrontRepository) ListNewArrivalProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	var products []*models.Product
	err := r.preloadProducts(r.db.WithContext(ctx)).
		Where("status = ?", constants.ProductStatusActive).
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

type wishlistRankRow struct {
	ProductID uint
	LikeCount int64
}

// ListMostWishlistedIDs ranks products by wishlist count.
func (r *StorefrontRepository) ListMostWishlistedIDs(ctx context.Context, limit int) ([]wishlistRankRow, error) {
	var rows []wishlistRankRow
	err := r.db.WithContext(ctx).Table("product_likes").
		Select("product_id, COUNT(*) AS like_count").
		Joins("INNER JOIN products p ON p.id = product_likes.product_id AND p.deleted_at IS NULL AND p.status = ?", constants.ProductStatusActive).
		Group("product_id").
		Order("like_count DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ListCustomerFavoriteProducts returns highest-rated products with reviews.
func (r *StorefrontRepository) ListCustomerFavoriteProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	var products []*models.Product
	err := r.preloadProducts(r.db.WithContext(ctx)).
		Where("status = ? AND reviews_count > 0", constants.ProductStatusActive).
		Order("rating DESC, reviews_count DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

type featuredStoreRow struct {
	models.Store
	ProductCount int64   `gorm:"column:product_count"`
	UnitsSold    int64   `gorm:"column:units_sold"`
	Revenue      float64 `gorm:"column:revenue"`
}

// ListFeaturedStores ranks stores by revenue from recent orders.
func (r *StorefrontRepository) ListFeaturedStores(ctx context.Context, limit int) ([]featuredStoreRow, error) {
	since := time.Now().AddDate(0, -6, 0)
	var rows []featuredStoreRow
	err := r.db.WithContext(ctx).Table("stores").
		Select(`
			stores.*,
			COUNT(DISTINCT products.id) AS product_count,
			COALESCE(SUM(order_items.quantity), 0) AS units_sold,
			COALESCE(SUM(order_items.quantity * order_items.price), 0) AS revenue`).
		Joins("INNER JOIN products ON products.store_id = stores.id AND products.deleted_at IS NULL AND products.status = ?", constants.ProductStatusActive).
		Joins("LEFT JOIN order_items ON order_items.product_id = products.id").
		Joins("LEFT JOIN orders o ON o.id = order_items.order_id AND o.deleted_at IS NULL AND o.created_at >= ? AND o.status IN ?", since, revenueOrderStatuses).
		Where("stores.deleted_at IS NULL AND stores.status = ?", "active").
		Group("stores.id").
		Order("revenue DESC, stores.rating DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

// ListCategoriesByIDs loads active categories preserving the given id order.
func (r *StorefrontRepository) ListCategoriesByIDs(ctx context.Context, ids []uint) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}
	var categories []models.Category
	if err := r.db.WithContext(ctx).
		Where("id IN ? AND is_active = ? AND deleted_at IS NULL", ids, true).
		Find(&categories).Error; err != nil {
		return nil, err
	}
	byID := make(map[uint]models.Category, len(categories))
	for _, c := range categories {
		byID[c.ID] = c
	}
	ordered := make([]models.Category, 0, len(ids))
	for _, id := range ids {
		if c, ok := byID[id]; ok {
			ordered = append(ordered, c)
		}
	}
	return ordered, nil
}

// ListProductsByCategory returns active products in a category ordered by rating.
func (r *StorefrontRepository) ListProductsByCategory(ctx context.Context, categoryID uint, limit int) ([]*models.Product, error) {
	var products []*models.Product
	err := r.preloadProducts(r.db.WithContext(ctx)).
		Where("status = ? AND category_id = ?", constants.ProductStatusActive, categoryID).
		Order("rating DESC, reviews_count DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// ListProductsByMaxPrice returns active products under a price ceiling.
func (r *StorefrontRepository) ListProductsByMaxPrice(ctx context.Context, maxPrice float64, limit int) ([]*models.Product, error) {
	var products []*models.Product
	err := r.preloadProducts(r.db.WithContext(ctx)).
		Where("status = ? AND price <= ?", constants.ProductStatusActive, maxPrice).
		Order("rating DESC").
		Limit(limit).
		Find(&products).Error
	return products, err
}

// ListRecentlyRestockedProductIDs finds products restocked in the last 30 days.
func (r *StorefrontRepository) ListRecentlyRestockedProductIDs(ctx context.Context, limit int) ([]uint, error) {
	since := time.Now().AddDate(0, 0, -30)
	var adjustments []models.InventoryAdjustment
	err := r.db.WithContext(ctx).
		Where("created_at >= ? AND adjustment_type IN ?", since, []string{constants.InventoryAdjReturnRestock, constants.InventoryAdjReceive, constants.InventoryAdjCorrection}).
		Order("created_at DESC").
		Limit(limit * 3).
		Find(&adjustments).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[uint]struct{})
	ids := make([]uint, 0, limit)
	for _, adj := range adjustments {
		if _, ok := seen[adj.ProductID]; ok {
			continue
		}
		seen[adj.ProductID] = struct{}{}
		ids = append(ids, adj.ProductID)
		if len(ids) >= limit {
			break
		}
	}
	return ids, nil
}

// FindBrandBannerImages returns a representative product image per brand.
func (r *StorefrontRepository) FindBrandBannerImages(ctx context.Context, brandIDs []uint) (map[uint]string, error) {
	out := make(map[uint]string, len(brandIDs))
	if len(brandIDs) == 0 {
		return out, nil
	}
	var products []models.Product
	err := r.db.WithContext(ctx).
		Select("brand_id", "images").
		Where("brand_id IN ?", brandIDs).
		Where("status = ?", constants.ProductStatusActive).
		Where("array_length(images, 1) > 0").
		Order("rating DESC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		if p.BrandID == nil {
			continue
		}
		if _, ok := out[*p.BrandID]; ok {
			continue
		}
		if len(p.Images) > 0 {
			out[*p.BrandID] = p.Images[0]
		}
	}
	return out, nil
}
