package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// OrderListQuery filters order list queries at the persistence layer.
type OrderListQuery struct {
	UserID         *uint
	StoreID        *uint
	Status         string
	PaymentStatus  string
	ShipmentStatus string
	Tag            string
	Search         string
	FromDate       *time.Time
	ToDate         *time.Time
	MinAmount      *float64
	MaxAmount      *float64
	Limit          int
	Offset         int
	PreloadUser    bool
}

// OrderRepository implements order persistence with GORM.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a GORM-backed order repository.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// FindByIDAndUserID loads an order scoped to a user with relations.
func (r *OrderRepository) FindByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByID loads an order by id with optional user preload.
func (r *OrderRepository) FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error) {
	q := r.db.WithContext(ctx)
	if preloadUser {
		q = q.Preload("User")
	}
	var order models.Order
	if err := q.First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// FindAdminByID loads an order with admin detail preloads.
func (r *OrderRepository) FindAdminByID(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("WorkflowState").
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Order("tag ASC")
		}).
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Preload("Payment").
		Preload("Shipment").
		First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) applyStoreScope(query *gorm.DB, storeID uint) *gorm.DB {
	return query.
		Joins("JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Joins("JOIN products ON products.id = order_items.product_id AND products.deleted_at IS NULL").
		Where("products.store_id = ?", storeID)
}

func (r *OrderRepository) applyListFilters(query *gorm.DB, q OrderListQuery) *gorm.DB {
	if q.StoreID != nil {
		query = r.applyStoreScope(query, *q.StoreID)
	}
	if q.UserID != nil {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Status != "" {
		query = query.Where("orders.status = ?", q.Status)
	}
	if q.PaymentStatus != "" {
		query = query.Joins("LEFT JOIN payments ON payments.order_id = orders.id AND payments.deleted_at IS NULL").
			Where("payments.status = ?", q.PaymentStatus)
	}
	if q.ShipmentStatus != "" {
		query = query.Joins("LEFT JOIN shipments ON shipments.order_id = orders.id AND shipments.deleted_at IS NULL").
			Where("shipments.status = ?", q.ShipmentStatus)
	}
	if q.Tag != "" {
		query = query.Joins("JOIN order_tags ON order_tags.order_id = orders.id").
			Where("order_tags.tag = ?", q.Tag)
	}
	if q.FromDate != nil {
		query = query.Where("created_at >= ?", q.FromDate)
	}
	if q.ToDate != nil {
		query = query.Where("created_at <= ?", q.ToDate)
	}
	if q.MinAmount != nil {
		query = query.Where("total_amount >= ?", *q.MinAmount)
	}
	if q.MaxAmount != nil {
		query = query.Where("total_amount <= ?", *q.MaxAmount)
	}
	if q.Search != "" {
		term := "%" + q.Search + "%"
		query = query.Joins("User").
			Where(
				"orders.order_number ILIKE ? OR users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?",
				term, term, term, term,
			)
	}
	return query
}

// UserIDsWithProductPurchase returns user IDs from the given set who bought the product
// in a paid, shipped, or delivered order.
func (r *OrderRepository) UserIDsWithProductPurchase(
	ctx context.Context,
	productID uint,
	userIDs []uint,
) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if productID == 0 || len(userIDs) == 0 {
		return result, nil
	}

	var purchased []uint
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Distinct("orders.user_id").
		Joins("JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Where("order_items.product_id = ?", productID).
		Where("orders.user_id IN ?", userIDs).
		Where("orders.status IN ?", []string{
			constants.OrderStatusPaid,
			constants.OrderStatusShipped,
			constants.OrderStatusDelivered,
		}).
		Pluck("orders.user_id", &purchased).Error
	if err != nil {
		return nil, err
	}

	for _, id := range purchased {
		result[id] = true
	}
	return result, nil
}
