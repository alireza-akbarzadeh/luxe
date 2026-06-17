package services

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type AuditServiceInterface interface {
	Log(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error)
}

type auditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditServiceInterface {
	return &auditService{db: db}
}

func (s *auditService) Log(ctx context.Context, entry *models.AuditLog) error {
	if err := s.db.WithContext(ctx).Create(entry).Error; err != nil {
		utils.Log.WithError(err).Warn("failed to write audit log")
		return err
	}
	return nil
}

func (s *auditService) List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	q := s.db.WithContext(ctx).Model(&models.AuditLog{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, total, err
}
