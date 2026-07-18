package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// StoreHolidayRepository persists store holidays and their store scoping with GORM.
type StoreHolidayRepository struct {
	db *gorm.DB
}

// NewStoreHolidayRepository creates a GORM-backed store holiday repository.
func NewStoreHolidayRepository(db *gorm.DB) *StoreHolidayRepository {
	return &StoreHolidayRepository{db: db}
}

// GetByID loads a holiday with its scoped stores preloaded.
func (r *StoreHolidayRepository) GetByID(ctx context.Context, id uint) (*models.StoreHoliday, error) {
	var holiday models.StoreHoliday
	if err := r.db.WithContext(ctx).Preload("Stores").First(&holiday, id).Error; err != nil {
		return nil, err
	}
	return &holiday, nil
}

// Create inserts a holiday row and its store scoping.
func (r *StoreHolidayRepository) Create(ctx context.Context, holiday *models.StoreHoliday, storeIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(holiday).Error; err != nil {
			return err
		}
		return replaceHolidayStores(tx, holiday.ID, storeIDs)
	})
}

// Save persists holiday field changes and replaces store scoping.
func (r *StoreHolidayRepository) Save(ctx context.Context, holiday *models.StoreHoliday, storeIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(holiday).Error; err != nil {
			return err
		}
		return replaceHolidayStores(tx, holiday.ID, storeIDs)
	})
}

func replaceHolidayStores(tx *gorm.DB, holidayID uint, storeIDs []uint) error {
	if err := tx.Where("holiday_id = ?", holidayID).Delete(&models.StoreHolidayStore{}).Error; err != nil {
		return err
	}
	if len(storeIDs) == 0 {
		return nil
	}
	rows := make([]models.StoreHolidayStore, 0, len(storeIDs))
	for _, id := range storeIDs {
		rows = append(rows, models.StoreHolidayStore{HolidayID: holidayID, StoreID: id})
	}
	return tx.Create(&rows).Error
}

// DeleteByID soft-deletes a holiday.
func (r *StoreHolidayRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.StoreHoliday{}, id)
	return result.RowsAffected, result.Error
}

// List returns paginated holidays matching filters.
func (r *StoreHolidayRepository) List(ctx context.Context, req *dto.ListStoreHolidaysRequest) ([]models.StoreHoliday, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.StoreHoliday{})

	if req.Search != "" {
		search := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(store_holidays.name) LIKE ? OR LOWER(store_holidays.description) LIKE ?", search, search)
	}
	if req.HolidayType != "" {
		query = query.Where("store_holidays.holiday_type = ?", req.HolidayType)
	}
	if req.Status != "" {
		query = query.Where("store_holidays.status = ?", req.Status)
	}
	if req.Region != "" {
		query = query.Where("store_holidays.region = ?", req.Region)
	}
	if req.StoreID > 0 {
		query = query.Where(
			"store_holidays.apply_to = 'all' OR EXISTS (SELECT 1 FROM store_holiday_stores shs WHERE shs.holiday_id = store_holidays.id AND shs.store_id = ?)",
			req.StoreID,
		)
	}
	if req.Year > 0 && req.Month > 0 {
		monthStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)
		query = query.Where("store_holidays.start_date < ? AND store_holidays.end_date >= ?", monthEnd, monthStart)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	var holidays []models.StoreHoliday
	err := query.Preload("Stores").
		Order("store_holidays.start_date ASC").
		Offset(offset).
		Limit(limit).
		Find(&holidays).Error
	if err != nil {
		return nil, 0, err
	}
	return holidays, total, nil
}

// ListForRange returns published holidays overlapping [start, end), scoped to a store/region when provided.
func (r *StoreHolidayRepository) ListForRange(ctx context.Context, start, end time.Time, storeID *uint, region *string) ([]models.StoreHoliday, error) {
	query := r.db.WithContext(ctx).Model(&models.StoreHoliday{}).
		Where("status = ?", "published").
		Where("start_date < ? AND end_date >= ?", end, start)

	scopeConditions := []string{"apply_to = 'all'"}
	scopeArgs := make([]interface{}, 0, 2)
	if region != nil && *region != "" {
		scopeConditions = append(scopeConditions, "(apply_to = 'region' AND region = ?)")
		scopeArgs = append(scopeArgs, *region)
	}
	if storeID != nil {
		scopeConditions = append(scopeConditions, "EXISTS (SELECT 1 FROM store_holiday_stores shs WHERE shs.holiday_id = store_holidays.id AND shs.store_id = ?)")
		scopeArgs = append(scopeArgs, *storeID)
	}
	query = query.Where(strings.Join(scopeConditions, " OR "), scopeArgs...)

	var holidays []models.StoreHoliday
	err := query.Order("priority DESC, start_date ASC").Find(&holidays).Error
	return holidays, err
}

// ListUpcoming returns published holidays starting from a date, limited.
func (r *StoreHolidayRepository) ListUpcoming(ctx context.Context, from time.Time, limit int) ([]models.StoreHoliday, error) {
	var holidays []models.StoreHoliday
	err := r.db.WithContext(ctx).Model(&models.StoreHoliday{}).
		Where("status = ? AND end_date >= ?", "published", from).
		Order("start_date ASC").
		Limit(limit).
		Find(&holidays).Error
	return holidays, err
}

// CountUpcoming counts published holidays ending on/after a date.
func (r *StoreHolidayRepository) CountUpcoming(ctx context.Context, from time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.StoreHoliday{}).
		Where("status = ? AND end_date >= ?", "published", from).
		Count(&count).Error
	return count, err
}

// SetStatus updates a holiday's status (used for publish).
func (r *StoreHolidayRepository) SetStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.StoreHoliday{}).Where("id = ?", id).Update("status", status).Error
}

// StoreIDsForHoliday returns the scoped store ids for a holiday.
func (r *StoreHolidayRepository) StoreIDsForHoliday(ctx context.Context, holidayID uint) ([]uint, error) {
	var storeIDs []uint
	err := r.db.WithContext(ctx).Model(&models.StoreHolidayStore{}).
		Where("holiday_id = ?", holidayID).
		Pluck("store_id", &storeIDs).Error
	return storeIDs, err
}
