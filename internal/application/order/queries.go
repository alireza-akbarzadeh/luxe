package order

import (
	"context"
	"errors"

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
