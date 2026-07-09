package order

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ListFilter mirrors admin/user order list filters at the application layer.
type ListFilter struct {
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

// VendorOrderStats aggregates order counts for a vendor store.
type VendorOrderStats struct {
	Total     int64            `json:"total"`
	ByStatus  map[string]int64 `json:"by_status"`
}

// Reader loads order models for HTTP handlers and application use cases.
type Reader interface {
	List(ctx context.Context, filter ListFilter) ([]models.Order, int64, error)
	FindByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error)
	FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error)
	FindAdminByID(ctx context.Context, orderID uint) (*models.Order, error)
	CountByStoreStatus(ctx context.Context, storeID uint) (VendorOrderStats, error)
	OrderBelongsToStore(ctx context.Context, orderID, storeID uint) (bool, error)
	GetVendorStoreSalesSummary(ctx context.Context, storeID uint, from, to time.Time) (VendorSalesSummary, error)
	ListVendorTopProducts(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]VendorTopProduct, error)
	ListVendorDailySales(ctx context.Context, storeID uint, from, to time.Time) ([]VendorDailySales, error)
	ListVendorStoreCustomers(ctx context.Context, storeID uint, from, to time.Time, limit int) ([]VendorStoreCustomer, error)
}

// Writer persists order status changes.
type Writer interface {
	UpdateStatus(ctx context.Context, orderID uint, status string) error
	UpdateStatusByIDs(ctx context.Context, orderIDs []uint, status string) (int64, error)
	FindOverduePaid(ctx context.Context, cutoff time.Time, excludedStatuses []string) ([]models.Order, error)
	Save(ctx context.Context, order *models.Order) error
	UpdateNotes(ctx context.Context, orderID uint, notes string) error
	ReplaceTags(ctx context.Context, orderID uint, tags []string) error
	UpdateShipmentByOrderID(ctx context.Context, orderID uint, updates map[string]interface{}) error
}

// ListFilterFromDTO maps storefront order filters.
func ListFilterFromDTO(userID uint, filters dto.OrderListFilters) ListFilter {
	limit := filters.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return ListFilter{
		UserID:    &userID,
		Status:    filters.Status,
		FromDate:  filters.FromDate,
		ToDate:    filters.ToDate,
		MinAmount: filters.MinAmount,
		MaxAmount: filters.MaxAmount,
		Limit:     limit,
		Offset:    filters.Offset,
	}
}