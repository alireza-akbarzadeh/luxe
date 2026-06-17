package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/repositories"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
)

type AuditServiceInterface interface {
	Log(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error)
}

type auditService struct {
	repo repositories.AuditRepository
}

func NewAuditService(repo repositories.AuditRepository) AuditServiceInterface {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, entry *models.AuditLog) error {
	if err := s.repo.Create(ctx, entry); err != nil {
		utils.Log.WithError(err).Warn("failed to write audit log")
		return err
	}
	return nil
}

func (s *auditService) List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error) {
	return s.repo.List(ctx, limit, offset)
}
