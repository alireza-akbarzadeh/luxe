// Package services provides the implementation of the admin service.
package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type AdminServiceInterface interface {
	GetStats(ctx context.Context) (*dto.AdminStatsResponse, error)
	GetDashboardOverview(ctx context.Context, filters dto.AdminDashboardFilters) (*dto.AdminDashboardOverviewResponse, error)
	GetRevenueReport(ctx context.Context, filters dto.AdminRevenueReportFilters) (*dto.AdminRevenueReportResponse, error)
	GetSalesFeedSnapshot(ctx context.Context) (*dto.AdminSalesFeedSnapshotResponse, error)
	ListUsers(ctx context.Context, filters dto.AdminUserFilters) ([]dto.AdminUserResponse, int64, error)
	UpdateUserRole(ctx context.Context, userID uint, role string) error
	ToggleUserActive(ctx context.Context, userID uint, active bool) error
	ExportOrdersCSV(ctx context.Context, filters dto.AdminOrderExportFilters) ([]byte, error)
}

type adminService struct {
	db     *gorm.DB
	engine *workflow.Engine
	roles  RoleServiceInterface
}

func NewAdminService(db *gorm.DB, engine *workflow.Engine, roles RoleServiceInterface) AdminServiceInterface {
	return &adminService{db: db, engine: engine, roles: roles}
}

func (s *adminService) GetStats(ctx context.Context) (*dto.AdminStatsResponse, error) {
	db := s.db.WithContext(ctx)

	var totalUsers int64
	if err := db.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var activeUsers int64
	if err := db.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var adminUsers int64
	if err := db.Model(&models.User{}).Where("role = ?", constants.RoleAdmin).Count(&adminUsers).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalOrders int64
	if err := db.Model(&models.Order{}).Count(&totalOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalProducts int64
	if err := db.Model(&models.Product{}).Where("status = ?", constants.ProductStatusActive).Count(&totalProducts).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalRevenue float64
	if err := db.Model(&models.Order{}).
		Where("status IN ?", []string{constants.OrderStatusPaid, constants.OrderStatusShipped, constants.OrderStatusDelivered}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&totalRevenue).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var pendingOrders int64
	if err := db.Model(&models.Order{}).Where("status = ?", constants.OrderStatusPending).Count(&pendingOrders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var totalWalletBalance float64
	if err := db.Model(&models.Wallet{}).
		Select("COALESCE(SUM(balance), 0)").
		Scan(&totalWalletBalance).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var lowStockProducts int64
	if err := db.Model(&models.Product{}).
		Where("stock <= low_stock_threshold AND status = ?", constants.ProductStatusActive).
		Count(&lowStockProducts).Error; err != nil {
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

func (s *adminService) ListUsers(ctx context.Context, filters dto.AdminUserFilters) ([]dto.AdminUserResponse, int64, error) {
	db := s.db.WithContext(ctx).Model(&models.User{})

	if filters.Search != "" {
		term := "%" + strings.ToLower(filters.Search) + "%"
		db = db.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(CONCAT(first_name, ' ', last_name)) LIKE ?",
			term, term, term, term,
		)
	} else if filters.Email != "" {
		db = db.Where("email ILIKE ?", "%"+filters.Email+"%")
	}
	if filters.Role != "" {
		db = db.Where("role = ?", filters.Role)
	}
	if filters.IsActive != nil {
		db = db.Where("is_active = ?", *filters.IsActive)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var users []models.User
	if err := db.Limit(filters.Limit).Offset(filters.Offset).
		Order("created_at DESC").Find(&users).Error; err != nil {
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

func (s *adminService) UpdateUserRole(ctx context.Context, userID uint, role string) error {
	exists, err := s.roles.RoleSlugExists(ctx, role)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrBadRequest("invalid role: role does not exist")
	}
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("role", role)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("user not found")
	}
	return nil
}

func (s *adminService) ExportOrdersCSV(ctx context.Context, filters dto.AdminOrderExportFilters) ([]byte, error) {
	db := s.db.WithContext(ctx).Model(&models.Order{}).
		Preload("User").
		Order("created_at DESC")

	if filters.Status != "" {
		db = db.Where("status = ?", filters.Status)
	}
	if filters.FromDate != "" {
		if t, err := time.Parse(time.DateOnly, filters.FromDate); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if filters.ToDate != "" {
		if t, err := time.Parse(time.DateOnly, filters.ToDate); err == nil {
			db = db.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var orders []models.Order
	if err := db.Limit(10000).Find(&orders).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "order_number", "status", "total_amount", "currency", "user_email", "created_at"})
	for _, o := range orders {
		email := ""
		if o.User.Email != "" {
			email = o.User.Email
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", o.ID),
			o.OrderNumber,
			o.Status,
			fmt.Sprintf("%.2f", o.TotalAmount),
			o.Currency,
			email,
			o.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return buf.Bytes(), nil
}

func (s *adminService) ToggleUserActive(ctx context.Context, userID uint, active bool) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("is_active", active)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("user not found")
	}

	if active {
		syncUserWorkflowState(ctx, s.engine, userID, "active", "unblock", nil)
	} else {
		syncUserWorkflowState(ctx, s.engine, userID, "blocked", "block", nil)
	}
	return nil
}
