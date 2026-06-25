package admin

import (
	"context"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	approle "github.com/alireza-akbarzadeh/luxe/internal/application/role"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service orchestrates admin dashboard and user management use cases.
type Service struct {
	queries  *Queries
	commands *Commands
	engine   *workflow.Engine
	roles    *approle.Queries
}

// NewService wires admin queries and commands.
func NewService(db *gorm.DB, engine *workflow.Engine, roles *approle.Queries) *Service {
	repo := postgres.NewAdminRepository(db)
	return &Service{
		queries:  NewQueries(repo),
		commands: NewCommands(repo),
		engine:   engine,
		roles:    roles,
	}
}

func (s *Service) GetStats(ctx context.Context) (*dto.AdminStatsResponse, error) {
	return s.queries.GetStats(ctx)
}

func (s *Service) GetDashboardOverview(ctx context.Context, filters dto.AdminDashboardFilters) (*dto.AdminDashboardOverviewResponse, error) {
	return s.queries.GetDashboardOverview(ctx, filters)
}

func (s *Service) GetRevenueReport(ctx context.Context, filters dto.AdminRevenueReportFilters) (*dto.AdminRevenueReportResponse, error) {
	return s.queries.GetRevenueReport(ctx, filters)
}

func (s *Service) GetSalesFeedSnapshot(ctx context.Context) (*dto.AdminSalesFeedSnapshotResponse, error) {
	return s.queries.GetSalesFeedSnapshot(ctx)
}

func (s *Service) ListUsers(ctx context.Context, filters dto.AdminUserFilters) ([]dto.AdminUserResponse, int64, error) {
	return s.queries.ListUsers(ctx, filters)
}

func (s *Service) UpdateUserRole(ctx context.Context, userID uint, role string) error {
	exists, err := s.roles.RoleSlugExists(ctx, role)
	if err != nil {
		return err
	}
	if !exists {
		return utils.ErrBadRequest("invalid role: role does not exist")
	}
	return s.commands.UpdateUserRole(ctx, userID, role)
}

func (s *Service) ExportOrdersCSV(ctx context.Context, filters dto.AdminOrderExportFilters) ([]byte, error) {
	return s.commands.ExportOrdersCSV(ctx, filters)
}

func (s *Service) ToggleUserActive(ctx context.Context, userID uint, active bool) error {
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
