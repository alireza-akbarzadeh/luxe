package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// CouponRepository implements coupon persistence with GORM.
type CouponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a GORM-backed coupon repository.
func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// FindByID loads a coupon with workflow state preloaded.
func (r *CouponRepository) FindByID(ctx context.Context, id uint) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := r.db.WithContext(ctx).Preload("WorkflowState").First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// FindByCode loads a coupon model by code.
func (r *CouponRepository) FindByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var coupon models.Coupon
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// FindActiveByCode loads an active, in-window coupon by code.
func (r *CouponRepository) FindActiveByCode(ctx context.Context, code string, now time.Time) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Where("(start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)", now, now).
		First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

// ExistsByCode checks whether a coupon code is taken.
func (r *CouponRepository) ExistsByCode(ctx context.Context, code string, excludeID uint) (bool, error) {
	var existing models.Coupon
	q := r.db.WithContext(ctx).Where("code = ?", code)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CreateCoupon inserts a coupon with selected fields.
func (r *CouponRepository) CreateCoupon(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Select(
		"Code", "Description", "ApplicationType", "DiscountType", "DiscountValue",
		"MinimumOrderAmount", "MaxDiscountAmount", "UsageLimit",
		"StartDate", "EndDate", "IsActive", "Conditions",
		"BogoBuyQuantity", "BogoGetQuantity", "BogoGetDiscountPercent",
	).Create(coupon).Error
}

// UpdateCouponIsActive sets is_active on a coupon.
func (r *CouponRepository) UpdateCouponIsActive(ctx context.Context, couponID uint, isActive bool) error {
	return r.db.WithContext(ctx).Model(&models.Coupon{}).Where("id = ?", couponID).Update("is_active", isActive).Error
}

// SaveCoupon persists coupon changes.
func (r *CouponRepository) SaveCoupon(ctx context.Context, coupon *models.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

// DeleteCoupon soft-deletes a coupon.
func (r *CouponRepository) DeleteCoupon(ctx context.Context, id uint) (int64, error) {
	result := r.db.WithContext(ctx).Delete(&models.Coupon{}, id)
	return result.RowsAffected, result.Error
}

// CountCouponUsageByUser counts usages for a coupon/user pair.
func (r *CouponRepository) CountCouponUsageByUser(ctx context.Context, couponID, userID uint) (int64, error) {
	var usageCount int64
	err := r.db.WithContext(ctx).Model(&models.CouponUsage{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).Count(&usageCount).Error
	return usageCount, err
}

// IncrementUsedCountTx increments used_count inside a transaction.
func (r *CouponRepository) IncrementUsedCountTx(tx *gorm.DB, coupon *models.Coupon) error {
	return tx.Model(coupon).Update("used_count", gorm.Expr("used_count + 1")).Error
}

// CreateCouponUsageTx inserts a usage record inside a transaction.
func (r *CouponRepository) CreateCouponUsageTx(tx *gorm.DB, usage *models.CouponUsage) error {
	return tx.Create(usage).Error
}

// ApplyOrderDiscountTx subtracts discount from order total inside a transaction.
func (r *CouponRepository) ApplyOrderDiscountTx(tx *gorm.DB, orderID uint, discount float64) error {
	return tx.Model(&models.Order{}).Where("id = ?", orderID).
		Update("total_amount", gorm.Expr("total_amount - ?", discount)).Error
}

// ListPublic returns paginated usable coupons.
func (r *CouponRepository) ListPublic(ctx context.Context, filters dto.CouponListFilters, now time.Time) ([]models.Coupon, int64, error) {
	var coupons []models.Coupon
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Coupon{}).
		Where("used_count < usage_limit AND start_date <= ? AND end_date >= ?", now, now)

	if filters.Code != "" {
		query = query.Where("code LIKE ?", "%"+filters.Code+"%")
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}
	if filters.DiscountType != "" {
		query = query.Where("discount_type = ?", filters.DiscountType)
	}
	if filters.StartDate != nil {
		query = query.Where("start_date >= ?", filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("end_date <= ?", filters.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	if err := query.Limit(limit).Offset(filters.Offset).Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// ListAdmin returns paginated coupons for admin with status filters.
func (r *CouponRepository) ListAdmin(ctx context.Context, filters dto.AdminCouponListFilters, now time.Time) ([]models.Coupon, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&models.Coupon{})

	switch filters.Status {
	case "active":
		query = query.Where(
			`EXISTS (
				SELECT 1 FROM workflow_states ws
				WHERE ws.id = coupons.workflow_state_id AND ws.code = ?
			) OR (coupons.workflow_state_id IS NULL AND coupons.is_active = ?)`,
			"active", true,
		)
	case "inactive":
		query = query.Where(
			`EXISTS (
				SELECT 1 FROM workflow_states ws
				WHERE ws.id = coupons.workflow_state_id AND ws.code IN ?
			) OR (coupons.workflow_state_id IS NULL AND coupons.is_active = ?)`,
			[]string{"draft", "paused"}, false,
		)
	case "expired":
		query = query.Where(
			`EXISTS (
				SELECT 1 FROM workflow_states ws
				WHERE ws.id = coupons.workflow_state_id AND ws.code = ?
			) OR coupons.end_date < ?`,
			"expired", now,
		)
	case "exhausted":
		query = query.Where(
			`EXISTS (
				SELECT 1 FROM workflow_states ws
				WHERE ws.id = coupons.workflow_state_id AND ws.code = ?
			) OR (coupons.usage_limit > 0 AND coupons.used_count >= coupons.usage_limit)`,
			"exhausted",
		)
	case "", "all":
	default:
		return nil, 0, gorm.ErrInvalidData
	}

	if filters.Code != "" {
		query = query.Where("code ILIKE ?", "%"+filters.Code+"%")
	}
	if filters.DiscountType != "" {
		query = query.Where("discount_type = ?", filters.DiscountType)
	}
	if filters.ApplicationType != "" {
		query = query.Where("application_type = ?", filters.ApplicationType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var coupons []models.Coupon
	if err := query.Preload("WorkflowState").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// ListAvailableForUser returns coupons valid for a user and order total.
func (r *CouponRepository) ListAvailableForUser(ctx context.Context, userID uint, orderTotal float64, now time.Time) ([]models.Coupon, error) {
	var coupons []models.Coupon
	query := r.db.WithContext(ctx).Model(&models.Coupon{}).
		Where("is_active = ? AND used_count < usage_limit AND start_date <= ? AND end_date >= ?",
			true, now, now)
	query = query.Where("id NOT IN (?)",
		r.db.Model(&models.CouponUsage{}).
			Select("coupon_id").
			Where("user_id = ?", userID))
	if orderTotal > 0 {
		query = query.Where("minimum_order_amount <= ?", orderTotal)
	}
	if err := query.Order("discount_value DESC").Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

// RecordUsageTx increments used count and creates usage inside a transaction.
func (r *CouponRepository) RecordUsageTx(tx *gorm.DB, couponID, userID, orderID uint, discountAmount float64) error {
	if err := tx.Model(&models.Coupon{}).Where("id = ?", couponID).
		Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
		return err
	}
	usage := &models.CouponUsage{
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
	}
	return tx.Create(usage).Error
}

// DB returns the underlying GORM handle.
func (r *CouponRepository) DB() *gorm.DB {
	return r.db
}
