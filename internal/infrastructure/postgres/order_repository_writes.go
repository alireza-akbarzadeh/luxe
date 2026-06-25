package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// List implements apporder.Reader.
func (r *OrderRepository) List(ctx context.Context, filter apporder.ListFilter) ([]models.Order, int64, error) {
	q := OrderListQuery{
		UserID:      filter.UserID,
		Status:      filter.Status,
		Search:      filter.Search,
		FromDate:    filter.FromDate,
		ToDate:      filter.ToDate,
		MinAmount:   filter.MinAmount,
		MaxAmount:   filter.MaxAmount,
		Limit:       filter.Limit,
		Offset:      filter.Offset,
		PreloadUser: filter.PreloadUser,
	}
	return r.list(ctx, q)
}

func (r *OrderRepository) list(ctx context.Context, q OrderListQuery) ([]models.Order, int64, error) {
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

// UpdateStatus sets order.status for a single order.
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

// UpdateStatusByIDs sets order.status for multiple orders.
func (r *OrderRepository) UpdateStatusByIDs(ctx context.Context, orderIDs []uint, status string) (int64, error) {
	if len(orderIDs) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Model(&models.Order{}).Where("id IN ?", orderIDs).Update("status", status)
	return res.RowsAffected, res.Error
}

// FindOverduePaid returns paid orders past cutoff excluding terminal statuses.
func (r *OrderRepository) FindOverduePaid(ctx context.Context, cutoff time.Time, excludedStatuses []string) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", constants.OrderStatusPaid, cutoff).
		Not("status IN ?", excludedStatuses).
		Find(&orders).Error
	return orders, err
}

// Save persists order field changes.
func (r *OrderRepository) Save(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}
