package repositories

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// OrderListQuery is the repository-level filter for listing orders.
type OrderListQuery struct {
	UserID     *uint
	Status     string
	FromDate   *time.Time
	ToDate     *time.Time
	MinAmount  *float64
	MaxAmount  *float64
	Limit      int
	Offset     int
	PreloadUser bool
}

// OrderRepository handles order persistence.
type OrderRepository interface {
	List(ctx context.Context, q OrderListQuery) ([]models.Order, int64, error)
	FindByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error)
	FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error)
	Save(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, orderID uint, status string) error
	FindOverduePaid(ctx context.Context, cutoff time.Time) ([]models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) applyListFilters(query *gorm.DB, q OrderListQuery) *gorm.DB {
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
	return query
}

func (r *orderRepository) List(ctx context.Context, q OrderListQuery) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Order{}), q)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Order{}), q)
	listQuery = listQuery.Limit(q.Limit).Offset(q.Offset).
		Preload("Items.Product").
		Preload("Payment").
		Preload("Shipment").
		Order("created_at DESC")

	if q.PreloadUser {
		listQuery = listQuery.Preload("User")
	}

	if err := listQuery.Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) FindByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error) {
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

func (r *orderRepository) FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error) {
	q := r.db.WithContext(ctx)
	if preloadUser {
		q = q.Preload("User")
	}
	var order models.Order
	err := q.First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Save(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (r *orderRepository) FindOverduePaid(ctx context.Context, cutoff time.Time) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.OrderStatusPaid, cutoff).
		Not("status IN (?)", []string{constants.OrderStatusDelivered, constants.OrderStatusCancelled, constants.OrderStatusRefunded}).
		Find(&orders).Error
	return orders, err
}

// OrderFiltersFromDTO maps API filters to repository query fields.
func OrderFiltersFromDTO(userID uint, filters dto.OrderListFilters) OrderListQuery {
	return OrderListQuery{
		UserID:    &userID,
		Status:    filters.Status,
		FromDate:  filters.FromDate,
		ToDate:    filters.ToDate,
		MinAmount: filters.MinAmount,
		MaxAmount: filters.MaxAmount,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}
}
