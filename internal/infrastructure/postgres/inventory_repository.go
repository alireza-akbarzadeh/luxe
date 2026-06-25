package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// InventoryRepository persists inventory ledger and stock queries with GORM.
type InventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository creates a GORM-backed inventory repository.
func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// DB exposes the underlying connection for external transaction coordination.
func (r *InventoryRepository) DB() *gorm.DB {
	return r.db
}

func (r *InventoryRepository) conn(ctx context.Context, tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

// Transaction runs fn inside a database transaction.
func (r *InventoryRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// CountLowStockActive counts active tracked products at or below threshold.
func (r *InventoryRepository) CountLowStockActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("track_inventory = ? AND stock <= low_stock_threshold AND stock > 0 AND status = ?",
			true, constants.ProductStatusActive).
		Count(&count).Error
	return count, err
}

// CountOutOfStockActive counts active tracked products with zero stock.
func (r *InventoryRepository) CountOutOfStockActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("track_inventory = ? AND stock = 0 AND status = ?", true, constants.ProductStatusActive).
		Count(&count).Error
	return count, err
}

// CountNotTracked counts products without inventory tracking.
func (r *InventoryRepository) CountNotTracked(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("track_inventory = ?", false).Count(&count).Error
	return count, err
}

// CountTrackedSKUs counts products with inventory tracking enabled.
func (r *InventoryRepository) CountTrackedSKUs(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("track_inventory = ?", true).Count(&count).Error
	return count, err
}

// SumTrackedUnits sums stock for tracked products.
func (r *InventoryRepository) SumTrackedUnits(ctx context.Context) (int64, error) {
	type sumRow struct {
		Total int64
	}
	var sum sumRow
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("COALESCE(SUM(stock), 0) AS total").
		Where("track_inventory = ?", true).
		Scan(&sum).Error
	return sum.Total, err
}

// CountActiveWaitlist counts active stock notification subscriptions.
func (r *InventoryRepository) CountActiveWaitlist(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.StockNotification{}).
		Where("status = ?", constants.StockNotificationStatusActive).
		Count(&count).Error
	return count, err
}

// InventoryListRow is a product row enriched for inventory list views.
type InventoryListRow struct {
	models.Product
	WaitlistCount int64 `gorm:"column:waitlist_count"`
	UnitsSold30d  int64 `gorm:"column:units_sold_30d"`
}

func (r *InventoryRepository) inventoryListQuery(ctx context.Context, revenueStatuses []string) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.Product{}).
		Select(`
			products.*,
			COALESCE(waitlist.waitlist_count, 0) AS waitlist_count,
			COALESCE(sales.units_sold_30d, 0) AS units_sold_30d`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) AS waitlist_count
			FROM stock_notifications
			WHERE status = ?
			GROUP BY product_id
		) waitlist ON waitlist.product_id = products.id`, constants.StockNotificationStatusActive).
		Joins(`LEFT JOIN (
			SELECT oi.product_id, COALESCE(SUM(oi.quantity), 0) AS units_sold_30d
			FROM order_items oi
			INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL
			WHERE o.created_at >= ? AND o.status IN ?
			GROUP BY oi.product_id
		) sales ON sales.product_id = products.id`,
			time.Now().AddDate(0, 0, -30),
			revenueStatuses,
		)
}

// ListInventoryItems returns paginated inventory rows.
func (r *InventoryRepository) ListInventoryItems(
	ctx context.Context,
	req *dto.ListInventoryRequest,
	revenueStatuses []string,
) ([]InventoryListRow, int64, error) {
	query := r.inventoryListQuery(ctx, revenueStatuses)

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("products.name ILIKE ? OR products.sku ILIKE ?", search, search)
	}

	switch req.StockStatus {
	case constants.InventoryStockLow:
		query = query.Where("products.track_inventory = ? AND products.stock > 0 AND products.stock <= products.low_stock_threshold", true)
	case constants.InventoryStockOut:
		query = query.Where("products.track_inventory = ? AND products.stock = 0", true)
	case constants.InventoryStockHealthy:
		query = query.Where("products.track_inventory = ? AND products.stock > products.low_stock_threshold", true)
	case constants.InventoryStockNotTracked:
		query = query.Where("products.track_inventory = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch req.Sort {
	case "stock_asc":
		query = query.Order("products.stock ASC, products.name ASC")
	case "stock_desc":
		query = query.Order("products.stock DESC, products.name ASC")
	case "name_asc":
		query = query.Order("products.name ASC")
	case "waitlist_desc":
		query = query.Order("waitlist_count DESC, products.stock ASC")
	case "velocity_desc":
		query = query.Order("units_sold_30d DESC, products.stock ASC")
	default:
		query = query.Order("products.stock ASC, products.name ASC")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var rows []InventoryListRow
	if err := query.Offset(offset).Limit(limit).Preload("WorkflowState").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetInventoryItem loads a single inventory list row.
func (r *InventoryRepository) GetInventoryItem(ctx context.Context, productID uint, revenueStatuses []string) (*InventoryListRow, error) {
	var row InventoryListRow
	err := r.inventoryListQuery(ctx, revenueStatuses).
		Where("products.id = ?", productID).
		Preload("WorkflowState").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetProductForUpdate loads a product row with row lock.
func (r *InventoryRepository) GetProductForUpdate(ctx context.Context, tx *gorm.DB, productID uint) (*models.Product, error) {
	var product models.Product
	if err := r.conn(ctx, tx).Set("gorm:query_option", "FOR UPDATE").First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// GetProduct loads a product without locking.
func (r *InventoryRepository) GetProduct(ctx context.Context, tx *gorm.DB, productID uint) (*models.Product, error) {
	var product models.Product
	if err := r.conn(ctx, tx).First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// SaveProduct persists product changes.
func (r *InventoryRepository) SaveProduct(ctx context.Context, tx *gorm.DB, product *models.Product) error {
	return r.conn(ctx, tx).Save(product).Error
}

// CreateAdjustment inserts an inventory adjustment row.
func (r *InventoryRepository) CreateAdjustment(ctx context.Context, tx *gorm.DB, adj *models.InventoryAdjustment) error {
	return r.conn(ctx, tx).Create(adj).Error
}

// CountAdjustments counts adjustments matching filters.
func (r *InventoryRepository) CountAdjustments(ctx context.Context, productID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.InventoryAdjustment{}).Where("product_id = ?", productID).Count(&total).Error
	return total, err
}

// ListAdjustments returns paginated adjustments for a product.
func (r *InventoryRepository) ListAdjustments(ctx context.Context, productID uint, offset, limit int) ([]models.InventoryAdjustment, error) {
	var rows []models.InventoryAdjustment
	err := r.db.WithContext(ctx).Model(&models.InventoryAdjustment{}).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Preload("Product").
		Preload("Actor").
		Find(&rows).Error
	return rows, err
}

// ListRecentAdjustments returns the most recent adjustments globally.
func (r *InventoryRepository) ListRecentAdjustments(ctx context.Context, limit int) ([]models.InventoryAdjustment, error) {
	var rows []models.InventoryAdjustment
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Preload("Product").
		Preload("Actor").
		Find(&rows).Error
	return rows, err
}

// FindProductBySKU loads a product by SKU.
func (r *InventoryRepository) FindProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// FindLowStockTracked returns tracked active products at or below threshold.
func (r *InventoryRepository) FindLowStockTracked(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Where("track_inventory = ? AND status = ? AND stock <= low_stock_threshold", true, constants.ProductStatusActive).
		Order("stock ASC").
		Find(&products).Error
	return products, err
}

// FindAdmins loads admin users.
func (r *InventoryRepository) FindAdmins(ctx context.Context) ([]models.User, error) {
	var admins []models.User
	err := r.db.WithContext(ctx).Where("role = ?", constants.RoleAdmin).Find(&admins).Error
	return admins, err
}

// GetReturnWithOrder loads a return with order items.
func (r *InventoryRepository) GetReturnWithOrder(ctx context.Context, returnID uint) (*models.Return, error) {
	var ret models.Return
	if err := r.db.WithContext(ctx).Preload("Order.Items").First(&ret, returnID).Error; err != nil {
		return nil, err
	}
	return &ret, nil
}

// CountReturnRestockAdjustments counts existing restock adjustments for a return line.
func (r *InventoryRepository) CountReturnRestockAdjustments(ctx context.Context, tx *gorm.DB, returnID, productID uint) (int64, error) {
	var existing int64
	err := r.conn(ctx, tx).Model(&models.InventoryAdjustment{}).
		Where("reference_type = ? AND reference_id = ? AND product_id = ? AND adjustment_type = ?",
			constants.InventoryRefReturn, returnID, productID, constants.InventoryAdjReturnRestock).
		Count(&existing).Error
	return existing, err
}
