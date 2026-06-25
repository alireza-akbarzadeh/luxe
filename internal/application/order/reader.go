package order

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// ListFilter mirrors admin/user order list filters at the application layer.
type ListFilter struct {
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

// Reader loads order models for HTTP and legacy services.
type Reader interface {
	List(ctx context.Context, filter ListFilter) ([]models.Order, int64, error)
	FindByIDAndUserID(ctx context.Context, orderID, userID uint) (*models.Order, error)
	FindByID(ctx context.Context, orderID uint, preloadUser bool) (*models.Order, error)
	FindAdminByID(ctx context.Context, orderID uint) (*models.Order, error)
}

// Writer persists order status changes.
type Writer interface {
	UpdateStatus(ctx context.Context, orderID uint, status string) error
	UpdateStatusByIDs(ctx context.Context, orderIDs []uint, status string) (int64, error)
	FindOverduePaid(ctx context.Context, cutoff time.Time, excludedStatuses []string) ([]models.Order, error)
	Save(ctx context.Context, order *models.Order) error
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