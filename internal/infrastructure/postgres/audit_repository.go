package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// AuditRepository persists audit logs with GORM.
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a GORM-backed audit repository.
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create inserts an audit log entry.
func (r *AuditRepository) Create(ctx context.Context, entry *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// CountFiltered counts audit logs matching filters.
func (r *AuditRepository) CountFiltered(ctx context.Context, filters dto.AuditLogListFilters) (int64, error) {
	var total int64
	q := r.applyAuditFilters(r.db.WithContext(ctx).Model(&models.AuditLog{}), filters)
	err := q.Count(&total).Error
	return total, err
}

// ListFiltered returns paginated audit logs matching filters.
func (r *AuditRepository) ListFiltered(ctx context.Context, filters dto.AuditLogListFilters, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.applyAuditFilters(r.db.WithContext(ctx).Model(&models.AuditLog{}), filters).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(filters.Offset).
		Find(&logs).Error
	return logs, err
}

// CountAll returns total audit log rows.
func (r *AuditRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&total).Error
	return total, err
}

// CountSince returns audit logs created on or after the given time.
func (r *AuditRepository) CountSince(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Where("created_at >= ?", since).
		Count(&count).Error
	return count, err
}

// CountDistinctUsers returns distinct user_id values in audit logs.
func (r *AuditRepository) CountDistinctUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AuditLog{}).
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

func (r *AuditRepository) applyAuditFilters(q *gorm.DB, filters dto.AuditLogListFilters) *gorm.DB {
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
