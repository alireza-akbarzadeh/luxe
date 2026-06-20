package services

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type AuditServiceInterface interface {
	Log(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, filters dto.AuditLogListFilters) ([]models.AuditLog, int64, error)
	Summary(ctx context.Context) (*dto.AuditLogSummaryResponse, error)
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

func (s *auditService) List(ctx context.Context, filters dto.AuditLogListFilters) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	q := s.applyAuditFilters(s.db.WithContext(ctx).Model(&models.AuditLog{}), filters)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filters.Limit
	if limit == 0 {
		limit = 20
	}

	err := s.applyAuditFilters(s.db.WithContext(ctx).Model(&models.AuditLog{}), filters).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(filters.Offset).
		Find(&logs).Error
	return logs, total, err
}

func (s *auditService) Summary(ctx context.Context) (*dto.AuditLogSummaryResponse, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	last24h := now.Add(-24 * time.Hour)

	var today int64
	if err := s.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at >= ?", startOfDay).
		Count(&today).Error; err != nil {
		return nil, err
	}

	var last24Hours int64
	if err := s.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at >= ?", last24h).
		Count(&last24Hours).Error; err != nil {
		return nil, err
	}

	var uniqueActors int64
	if err := s.db.WithContext(ctx).Model(&models.AuditLog{}).
		Distinct("user_id").
		Count(&uniqueActors).Error; err != nil {
		return nil, err
	}

	return &dto.AuditLogSummaryResponse{
		Total:        total,
		Last24Hours:  last24Hours,
		Today:        today,
		UniqueActors: uniqueActors,
	}, nil
}

func (s *auditService) applyAuditFilters(q *gorm.DB, filters dto.AuditLogListFilters) *gorm.DB {
	if filters.Action != "" {
		q = q.Where("action = ?", strings.ToUpper(filters.Action))
	}
	if filters.Resource != "" {
		q = q.Where("resource ILIKE ?", "%"+filters.Resource+"%")
	}
	if filters.UserID > 0 {
		q = q.Where("user_id = ?", filters.UserID)
	}
	if filters.Search != "" {
		term := "%" + strings.TrimSpace(filters.Search) + "%"
		q = q.Where(
			"path ILIKE ? OR resource ILIKE ? OR resource_id ILIKE ? OR EXISTS (SELECT 1 FROM users u WHERE u.id = audit_logs.user_id AND u.email ILIKE ?)",
			term, term, term, term,
		)
	}
	if from := parseAuditDate(filters.DateFrom); from != nil {
		q = q.Where("created_at >= ?", *from)
	}
	if to := parseAuditDate(filters.DateTo); to != nil {
		end := to.Add(24*time.Hour - time.Nanosecond)
		q = q.Where("created_at <= ?", end)
	}
	return q
}

func parseAuditDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			t := parsed.UTC()
			return &t
		}
	}
	return nil
}
