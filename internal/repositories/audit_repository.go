package repositories

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AuditRepository handles persistence for audit log entries.
type AuditRepository interface {
	Create(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, entry *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *auditRepository) List(ctx context.Context, limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	q := r.db.WithContext(ctx).Model(&models.AuditLog{})
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
