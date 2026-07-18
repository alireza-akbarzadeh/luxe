package admin

import (
	"context"
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates admin read use cases.
type Queries struct {
	repo    *postgres.AdminRepository
	navRepo *postgres.AdminNavRepository
}

// NewQueries creates admin query use cases.
func NewQueries(repo *postgres.AdminRepository, navRepo *postgres.AdminNavRepository) *Queries {
	return &Queries{repo: repo, navRepo: navRepo}
}

// GetStats returns platform-wide admin statistics.
func (q *Queries) GetStats(ctx context.Context) (*dto.AdminStatsResponse, error) {
	totalUsers, err := q.repo.CountUsers(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	activeUsers, err := q.repo.CountActiveUsers(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	adminUsers, err := q.repo.CountAdminUsers(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	totalOrders, err := q.repo.CountOrders(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	totalProducts, err := q.repo.CountActiveProducts(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	totalRevenue, err := q.repo.SumRevenue(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	pendingOrders, err := q.repo.CountPendingOrders(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	totalWalletBalance, err := q.repo.SumWalletBalance(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	lowStockProducts, err := q.repo.CountLowStockProducts(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.AdminStatsResponse{
		TotalUsers:          totalUsers,
		ActiveUsers:         activeUsers,
		AdminUsers:          adminUsers,
		TotalOrders:         totalOrders,
		TotalActiveProducts: totalProducts,
		TotalRevenue:        totalRevenue,
		PendingOrders:       pendingOrders,
		TotalWalletBalance:  totalWalletBalance,
		LowStockProducts:    lowStockProducts,
	}, nil
}

// ListUsers returns paginated users for admin.
func (q *Queries) ListUsers(ctx context.Context, filters dto.AdminUserFilters) ([]dto.AdminUserResponse, int64, error) {
	total, err := q.repo.CountUsersFiltered(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	users, err := q.repo.ListUsersFiltered(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	userIDs := make([]uint, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}
	statsByUser, err := q.repo.GetUserOrderStatsBatch(ctx, userIDs)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	result := make([]dto.AdminUserResponse, len(users))
	for i, u := range users {
		stats := statsByUser[u.ID]
		result[i] = dto.ToAdminUserResponse(&u, dto.UserOrderStats{
			OrderCount: stats.OrderCount,
			TotalSpent: stats.TotalSpent,
		})
	}
	return result, total, nil
}

// GetCustomerDetail returns a full customer profile with purchase stats.
func (q *Queries) GetCustomerDetail(ctx context.Context, userID uint) (*dto.AdminCustomerDetailResponse, error) {
	user, err := q.repo.FindUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("user not found")
		}
		return nil, utils.ErrInternal(err)
	}

	stats, err := q.repo.GetUserOrderStats(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	addressCount, err := q.repo.CountUserAddresses(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	base := dto.ToAdminUserResponse(user, dto.UserOrderStats{
		OrderCount: stats.OrderCount,
		TotalSpent: stats.TotalSpent,
	})

	return &dto.AdminCustomerDetailResponse{
		AdminUserResponse: base,
		PlusSubscribedAt:  user.PlusSubscribedAt,
		PlusExpiresAt:     user.PlusExpiresAt,
		AdminNotes:        user.AdminNotes,
		AddressCount:      addressCount,
	}, nil
}

// ListCustomerAddresses returns saved addresses for a customer (admin).
func (q *Queries) ListCustomerAddresses(ctx context.Context, userID uint) ([]models.Address, error) {
	if _, err := q.repo.FindUserByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("user not found")
		}
		return nil, utils.ErrInternal(err)
	}
	addresses, err := q.repo.ListUserAddresses(ctx, userID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return addresses, nil
}

// GetCustomerStats returns aggregate CRM metrics.
func (q *Queries) GetCustomerStats(ctx context.Context) (*dto.AdminCustomerStats, error) {
	total, err := q.repo.CountCustomers(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	plus, err := q.repo.CountPlusMembers(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	monthStart := time.Now().UTC().AddDate(0, -1, 0)
	newMonth, err := q.repo.CountNewCustomersSince(ctx, monthStart)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	vip, err := q.repo.CountCustomersBySegment(ctx, constants.CustomerSegmentVIP)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.AdminCustomerStats{
		TotalCustomers: total,
		PlusMembers:    plus,
		NewThisMonth:   newMonth,
		VipCustomers:   vip,
	}, nil
}
