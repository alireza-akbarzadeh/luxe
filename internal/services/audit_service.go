package services

import (
	"context"

	appaudit "github.com/alireza-akbarzadeh/luxe/internal/application/audit"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type AuditServiceInterface interface {
	Log(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, filters dto.AuditLogListFilters) ([]models.AuditLog, int64, error)
	Summary(ctx context.Context) (*dto.AuditLogSummaryResponse, error)
}

type auditService struct {
	commands *appaudit.Commands
	queries  *appaudit.Queries
}

func NewAuditService(db *gorm.DB) AuditServiceInterface {
	repo := postgres.NewAuditRepository(db)
	return &auditService{
		commands: appaudit.NewCommands(repo),
		queries:  appaudit.NewQueries(repo),
	}
}

func (s *auditService) Log(ctx context.Context, entry *models.AuditLog) error {
	return s.commands.Log(ctx, entry)
}

func (s *auditService) List(ctx context.Context, filters dto.AuditLogListFilters) ([]models.AuditLog, int64, error) {
	return s.queries.List(ctx, filters)
}

func (s *auditService) Summary(ctx context.Context) (*dto.AuditLogSummaryResponse, error) {
	return s.queries.Summary(ctx)
}
