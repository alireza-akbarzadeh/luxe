// Package services provides the implementation of the admin service.
package services

import (
	"context"

	appadmin "github.com/alireza-akbarzadeh/luxe/internal/application/admin"
	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
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
	queries  *appadmin.Queries
	commands *appadmin.Commands
	engine   *workflow.Engine
	roles    RoleServiceInterface
}

func NewAdminService(db *gorm.DB, engine *workflow.Engine, roles RoleServiceInterface) AdminServiceInterface {
	repo := postgres.NewAdminRepository(db)
	return &adminService{
		queries:  appadmin.NewQueries(repo),
		commands: appadmin.NewCommands(repo),
		engine:   engine,
		roles:    roles,
	}
}

func (s *adminService) GetStats(ctx context.Context) (*dto.AdminStatsResponse, error) {
	return s.queries.GetStats(ctx)
}

func (s *adminService) GetDashboardOverview(ctx context.Context, filters dto.AdminDashboardFilters) (*dto.AdminDashboardOverviewResponse, error) {
	return s.queries.GetDashboardOverview(ctx, filters)
}

func (s *adminService) GetRevenueReport(ctx context.Context, filters dto.AdminRevenueReportFilters) (*dto.AdminRevenueReportResponse, error) {
	return s.queries.GetRevenueReport(ctx, filters)
}

func (s *adminService) GetSalesFeedSnapshot(ctx context.Context) (*dto.AdminSalesFeedSnapshotResponse, error) {
	return s.queries.GetSalesFeedSnapshot(ctx)
}

func (s *adminService) ListUsers(ctx context.Context, filters dto.AdminUserFilters) ([]dto.AdminUserResponse, int64, error) {
	return s.queries.ListUsers(ctx, filters)
}

func (s *adminService) UpdateUserRole(ctx context.Context, userID uint, role string) error {
	exists, err := s.roles.RoleSlugExists(ctx, role)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrBadRequest("invalid role: role does not exist")
	}
	return s.commands.UpdateUserRole(ctx, userID, role)
}

func (s *adminService) ExportOrdersCSV(ctx context.Context, filters dto.AdminOrderExportFilters) ([]byte, error) {
	return s.commands.ExportOrdersCSV(ctx, filters)
}

func (s *adminService) ToggleUserActive(ctx context.Context, userID uint, active bool) error {
	if err := s.commands.ToggleUserActive(ctx, userID, active); err != nil {
		return err
	}

	if active {
		appworkflow.SyncUserState(ctx, s.engine, userID, "active", "unblock", nil)
	} else {
		appworkflow.SyncUserState(ctx, s.engine, userID, "blocked", "block", nil)
	}
	return nil
}
