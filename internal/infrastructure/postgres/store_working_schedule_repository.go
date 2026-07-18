package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// StoreWorkingScheduleRepository persists store working schedules with GORM.
type StoreWorkingScheduleRepository struct {
	db *gorm.DB
}

// NewStoreWorkingScheduleRepository creates a GORM-backed store working schedule repository.
func NewStoreWorkingScheduleRepository(db *gorm.DB) *StoreWorkingScheduleRepository {
	return &StoreWorkingScheduleRepository{db: db}
}

// GetByID loads a working schedule by id.
func (r *StoreWorkingScheduleRepository) GetByID(ctx context.Context, id uint) (*models.StoreWorkingSchedule, error) {
	var sched models.StoreWorkingSchedule
	if err := r.db.WithContext(ctx).First(&sched, id).Error; err != nil {
		return nil, err
	}
	return &sched, nil
}

// GetByStoreID loads the working schedule for a store, if any.
func (r *StoreWorkingScheduleRepository) GetByStoreID(ctx context.Context, storeID uint) (*models.StoreWorkingSchedule, error) {
	var sched models.StoreWorkingSchedule
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).First(&sched).Error
	if err != nil {
		return nil, err
	}
	return &sched, nil
}

// List returns working schedules, optionally filtered by store id.
func (r *StoreWorkingScheduleRepository) List(ctx context.Context, storeID *uint) ([]models.StoreWorkingSchedule, error) {
	query := r.db.WithContext(ctx).Model(&models.StoreWorkingSchedule{})
	if storeID != nil {
		query = query.Where("store_id = ?", *storeID)
	}
	var schedules []models.StoreWorkingSchedule
	err := query.Order("store_id ASC").Find(&schedules).Error
	return schedules, err
}

// Create inserts a working schedule row.
func (r *StoreWorkingScheduleRepository) Create(ctx context.Context, sched *models.StoreWorkingSchedule) error {
	return r.db.WithContext(ctx).Create(sched).Error
}

// Save persists working schedule field changes.
func (r *StoreWorkingScheduleRepository) Save(ctx context.Context, sched *models.StoreWorkingSchedule) error {
	return r.db.WithContext(ctx).Save(sched).Error
}

// CountDistinctStoresWithSchedule counts stores that have a working schedule row.
func (r *StoreWorkingScheduleRepository) CountDistinctStoresWithSchedule(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.StoreWorkingSchedule{}).Count(&count).Error
	return count, err
}
