package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// VendorOffDayRepository persists vendor-initiated off days with GORM.
type VendorOffDayRepository struct {
	db *gorm.DB
}

// NewVendorOffDayRepository creates a GORM-backed vendor off day repository.
func NewVendorOffDayRepository(db *gorm.DB) *VendorOffDayRepository {
	return &VendorOffDayRepository{db: db}
}

// GetByID loads a vendor off day by id.
func (r *VendorOffDayRepository) GetByID(ctx context.Context, id uint) (*models.VendorOffDay, error) {
	var offDay models.VendorOffDay
	if err := r.db.WithContext(ctx).First(&offDay, id).Error; err != nil {
		return nil, err
	}
	return &offDay, nil
}

// Create inserts a vendor off day row.
func (r *VendorOffDayRepository) Create(ctx context.Context, offDay *models.VendorOffDay) error {
	return r.db.WithContext(ctx).Create(offDay).Error
}

// Save persists vendor off day field changes.
func (r *VendorOffDayRepository) Save(ctx context.Context, offDay *models.VendorOffDay) error {
	return r.db.WithContext(ctx).Save(offDay).Error
}

// DeleteByID soft-deletes a vendor off day.
func (r *VendorOffDayRepository) DeleteByID(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.VendorOffDay{}, id)
	return result.RowsAffected, result.Error
}

// List returns paginated vendor off days matching filters.
func (r *VendorOffDayRepository) List(ctx context.Context, req *dto.ListVendorOffDaysRequest) ([]models.VendorOffDay, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.VendorOffDay{})

	if req.VendorID > 0 {
		query = query.Where("vendor_id = ?", req.VendorID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.OffType != "" {
		query = query.Where("off_type = ?", req.OffType)
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

	var offDays []models.VendorOffDay
	err := query.Order("start_date ASC").Offset(offset).Limit(limit).Find(&offDays).Error
	if err != nil {
		return nil, 0, err
	}
	return offDays, total, nil
}

// ListForRange returns published vendor off days overlapping [start, end), optionally scoped to a vendor.
func (r *VendorOffDayRepository) ListForRange(ctx context.Context, vendorID *uint, start, end time.Time) ([]models.VendorOffDay, error) {
	query := r.db.WithContext(ctx).Model(&models.VendorOffDay{}).
		Where("status = ?", "published").
		Where("start_date < ? AND end_date >= ?", end, start)
	if vendorID != nil {
		query = query.Where("vendor_id = ?", *vendorID)
	}
	var offDays []models.VendorOffDay
	err := query.Order("start_date ASC").Find(&offDays).Error
	return offDays, err
}

// ListUpcoming returns published off days ending on/after a date, limited.
func (r *VendorOffDayRepository) ListUpcoming(ctx context.Context, from time.Time, limit int) ([]models.VendorOffDay, error) {
	var offDays []models.VendorOffDay
	err := r.db.WithContext(ctx).Model(&models.VendorOffDay{}).
		Where("status = ? AND end_date >= ?", "published", from).
		Order("start_date ASC").
		Limit(limit).
		Find(&offDays).Error
	return offDays, err
}

// CountClosedToday counts published vendor off days covering the given day.
func (r *VendorOffDayRepository) CountClosedToday(ctx context.Context, day time.Time) (int64, error) {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VendorOffDay{}).
		Distinct("vendor_id").
		Where("status = ? AND start_date < ? AND end_date >= ?", "published", dayEnd, dayStart).
		Count(&count).Error
	return count, err
}
