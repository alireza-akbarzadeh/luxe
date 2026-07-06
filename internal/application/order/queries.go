package order

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates order read use cases.
type Queries struct {
	reader Reader
}

// NewQueries creates order query use cases.
func NewQueries(reader Reader) *Queries {
	return &Queries{reader: reader}
}

// ListUserOrders returns paginated orders for a customer.
func (q *Queries) ListUserOrders(ctx context.Context, userID uint, filters dto.OrderListFilters) ([]models.Order, int64, error) {
	orders, total, err := q.reader.List(ctx, ListFilterFromDTO(userID, filters))
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

// GetByIDForUser loads an order scoped to the owning user.
func (q *Queries) GetByIDForUser(ctx context.Context, orderID, userID uint) (*models.Order, error) {
	order, err := q.reader.FindByIDAndUserID(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

// GetAdmin loads an order with admin detail preloads.
func (q *Queries) GetAdmin(ctx context.Context, orderID uint) (*models.Order, error) {
	order, err := q.reader.FindAdminByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

// ListAdmin returns paginated orders for admin views.
func (q *Queries) ListAdmin(ctx context.Context, filter ListFilter) ([]models.Order, int64, error) {
	orders, total, err := q.reader.List(ctx, filter)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

// ListVendorStore returns paginated orders containing products from the given store.
func (q *Queries) ListVendorStore(ctx context.Context, storeID uint, filter ListFilter) ([]models.Order, int64, error) {
	filter.StoreID = &storeID
	filter.PreloadUser = true
	orders, total, err := q.reader.List(ctx, filter)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return orders, total, nil
}

// GetVendorStoreStats returns order counts for a vendor store.
func (q *Queries) GetVendorStoreStats(ctx context.Context, storeID uint) (VendorOrderStats, error) {
	stats, err := q.reader.CountByStoreStatus(ctx, storeID)
	if err != nil {
		return VendorOrderStats{}, utils.ErrInternal(err)
	}
	return stats, nil
}

// GetVendorStoreSalesSummary returns revenue metrics for a store in a date range.
func (q *Queries) GetVendorStoreSalesSummary(ctx context.Context, storeID uint, from, to time.Time) (VendorSalesSummary, error) {
	summary, err := q.reader.GetVendorStoreSalesSummary(ctx, storeID, from, to)
	if err != nil {
		return VendorSalesSummary{}, utils.ErrInternal(err)
	}
	return summary, nil
}

// ListVendorTopProducts returns top-selling products for a store in a date range.
func (q *Queries) ListVendorTopProducts(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]VendorTopProduct, error) {
	products, err := q.reader.ListVendorTopProducts(ctx, storeID, from, to, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return products, nil
}

// ListVendorDailySales returns per-day revenue for charting.
func (q *Queries) ListVendorDailySales(ctx context.Context, storeID uint, from, to time.Time) ([]VendorDailySales, error) {
	series, err := q.reader.ListVendorDailySales(ctx, storeID, from, to)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return series, nil
}

// ListVendorStoreCustomers returns buyer aggregates for a store in a date range.
func (q *Queries) ListVendorStoreCustomers(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]VendorStoreCustomer, error) {
	customers, err := q.reader.ListVendorStoreCustomers(ctx, storeID, from, to, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return customers, nil
}

// GetVendorStoreOrder loads an order if it belongs to the store.
func (q *Queries) GetVendorStoreOrder(ctx context.Context, storeID, orderID uint) (*models.Order, error) {
	ok, err := q.reader.OrderBelongsToStore(ctx, orderID, storeID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if !ok {
		return nil, utils.ErrNotFound("order not found")
	}
	order, err := q.reader.FindAdminByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}

// FindByID loads an order by id with optional user preload.
func (q *Queries) FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error) {
	order, err := q.reader.FindByID(ctx, orderID, preloadUser)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return order, nil
}
