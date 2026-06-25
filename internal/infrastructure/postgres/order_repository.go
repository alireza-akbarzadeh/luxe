package postgres

import (
	"context"
	"errors"
	"time"

	domainorder "github.com/alireza-akbarzadeh/luxe/internal/domain/order"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// OrderListQuery filters order list queries at the persistence layer.
type OrderListQuery struct {
	UserID      *uint
	Status      string
	Search      string
	FromDate    *time.Time
	ToDate      *time.Time
	MinAmount   *float64
	MaxAmount   *float64
	Limit       int
	Offset      int
	PreloadUser bool
}

// OrderRepository implements order persistence with GORM.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a GORM-backed order repository.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetByID implements domain/order.Repository.
func (r *OrderRepository) GetByID(ctx context.Context, id uint) (*domainorder.Order, error) {
	var m models.Order
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainorder.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomainOrder(&m), nil
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

func (r *OrderRepository) applyListFilters(query *gorm.DB, q OrderListQuery) *gorm.DB {
	if q.UserID != nil {
		query = query.Where("user_id = ?", *q.UserID)
	}
	if q.Status != "" {
		query = query.Where("status = ?", q.Status)
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

func toDomainOrder(m *models.Order) *domainorder.Order {
	return &domainorder.Order{
		ID:         m.ID,
		UserID:     m.UserID,
		Status:     m.Status,
		TotalCents: int64(m.TotalAmount * 100),
		Currency:   "USD",
	}
}
