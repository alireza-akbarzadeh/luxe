package admin

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates admin read use cases.
type Queries struct {
	repo *postgres.AdminRepository
}

// NewQueries creates admin query use cases.
func NewQueries(repo *postgres.AdminRepository) *Queries {
	return &Queries{repo: repo}
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

	result := make([]dto.AdminUserResponse, len(users))
	for i, u := range users {
		result[i] = dto.AdminUserResponse{
			ID:              u.ID,
			Email:           u.Email,
			FirstName:       u.FirstName,
			LastName:        u.LastName,
			Role:            u.Role,
			IsActive:        u.IsActive,
			EmailVerifiedAt: u.EmailVerifiedAt,
			LastLoginAt:     u.LastLoginAt,
			CreatedAt:       u.CreatedAt,
		}
	}
	return result, total, nil
}
