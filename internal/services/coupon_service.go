package services

import (
	"errors"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type CouponServiceInterface interface {
	Create(req dto.CreateCouponRequest) (*models.Coupon, error)
	GetByID(id uint) (*models.Coupon, error)
	GetByCode(code string) (*models.Coupon, error)
	Update(id uint, req dto.UpdateCouponRequest) (*models.Coupon, error)
	Delete(id uint) error
	List(dto.CouponListFilters) ([]models.Coupon, int64, error)
	ValidateCoupon(code string, userID uint, orderTotal float64) (*models.Coupon, float64, error)
	ApplyCoupon(tx *gorm.DB, userID uint, orderID uint, couponCode string, orderTotal float64) error
	GetAvailableCouponsForUser(userID uint, orderTotal float64) ([]models.Coupon, error)
}

type couponService struct {
	db *gorm.DB
}

func NewCouponService(db *gorm.DB) CouponServiceInterface {
	return &couponService{db: db}
}

// Create – store new coupon
func (s *couponService) Create(req dto.CreateCouponRequest) (*models.Coupon, error) {
	var existing models.Coupon
	if err := s.db.Where("code = ? ", req.Code).First(&existing).Error; err == nil {
		return nil, utils.ErrConflict("coupon code already exists")
	}

	coupon := &models.Coupon{
		Code:               req.Code,
		Description:        req.Description,
		DiscountType:       req.DiscountType,
		DiscountValue:      req.DiscountValue,
		MinimumOrderAmount: req.MinimumOrderAmount,
		MaxDiscountAmount:  req.MaxDiscountAmount,
		UsageLimit:         req.UsageLimit,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
	}
	if err := s.db.Create(coupon).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return coupon, nil
}

// ValidateCoupon checks if coupon is usable for a user and order total
func (s *couponService) ValidateCoupon(code string, userID uint, orderTotal float64) (*models.Coupon, float64, error) {
	var coupon models.Coupon
	now := time.Now()
	err := s.db.Where("code = ? AND is_active = ? AND start_date <= ? AND end_date >= ?", code, true, now, now).
		First(&coupon).Error
	if err != nil {
		return nil, 0, utils.ErrBadRequest("invalid or expired coupon")
	}
	if coupon.UsedCount >= coupon.UsageLimit {
		return nil, 0, utils.ErrBadRequest("coupon usage limit exceeded")
	}
	if orderTotal < coupon.MinimumOrderAmount {
		return nil, 0, utils.ErrBadRequest("order total below minimum amount")
	}
	// check if user already used this coupon (optional, can be allowed multiple times)
	var usageCount int64
	s.db.Model(&models.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", coupon.ID, userID).Count(&usageCount)
	if usageCount > 0 {
		return nil, 0, utils.ErrBadRequest("coupon already used by this user")
	}
	// calculate discount
	discount := 0.0
	if coupon.DiscountType == "percentage" {
		discount = orderTotal * (coupon.DiscountValue / 100)
		if coupon.MaxDiscountAmount != nil && discount > *coupon.MaxDiscountAmount {
			discount = *coupon.MaxDiscountAmount
		}
	} else { // fixed
		discount = coupon.DiscountValue
		if discount > orderTotal {
			discount = orderTotal
		}
	}
	return &coupon, discount, nil
}

// ApplyCoupon records usage and optionally updates order total (called during checkout)
func (s *couponService) ApplyCoupon(tx *gorm.DB, userID uint, orderID uint, couponCode string, orderTotal float64) error {
	coupon, discount, err := s.ValidateCoupon(couponCode, userID, orderTotal)
	if err != nil {
		return err
	}

	// start transaction
	// increment used_count
	if err := tx.Model(coupon).Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
		tx.Rollback()
		return utils.ErrInternal(err)
	}
	// create usage record
	usage := &models.CouponUsage{
		CouponID:       coupon.ID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discount,
	}
	if err := tx.Create(usage).Error; err != nil {
		tx.Rollback()
		return utils.ErrInternal(err)
	}
	// update order total (subtract discount)
	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("total_amount", gorm.Expr("total_amount - ?", discount)).Error; err != nil {
		tx.Rollback()
		return utils.ErrInternal(err)
	}
	return nil
}

// GetByID retrieves a coupon by its ID.

func (s *couponService) GetByID(couponID uint) (*models.Coupon, error) {
	var coupon models.Coupon
	err := s.db.First(&coupon, couponID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("coupon not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &coupon, nil
}

// GetByCode retrieves a coupon by its code.
func (s *couponService) GetByCode(code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := s.db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("coupon not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &coupon, nil
}

func (s *couponService) Update(id uint, req dto.UpdateCouponRequest) (*models.Coupon, error) {
	coupon, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	// Update fields only if provided
	if req.Code != nil && *req.Code != "" {
		var existing models.Coupon
		if err := s.db.Where("code = ? AND id != ?", *req.Code, id).First(&existing).Error; err == nil {
			return nil, utils.ErrConflict("coupon code already exists")
		}
		coupon.Code = *req.Code
	}
	if req.Description != nil {
		coupon.Description = *req.Description
	}
	if req.DiscountType != nil {
		coupon.DiscountType = *req.DiscountType
	}
	if req.DiscountValue != nil {
		coupon.DiscountValue = *req.DiscountValue
	}
	if req.MinimumOrderAmount != nil {
		coupon.MinimumOrderAmount = *req.MinimumOrderAmount
	}
	if req.MaxDiscountAmount != nil {
		coupon.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.UsageLimit != nil {
		coupon.UsageLimit = *req.UsageLimit
	}
	if req.StartDate != nil {
		coupon.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		coupon.EndDate = *req.EndDate
	}
	if req.IsActive != nil {
		coupon.IsActive = *req.IsActive
	}

	if err := s.db.Save(coupon).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return coupon, nil
}

// Delete soft-deletes a coupon.
func (s *couponService) Delete(id uint) error {
	result := s.db.Delete(&models.Coupon{}, id)
	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return utils.ErrNotFound("coupon not found")
	}
	return nil
}

// List returns paginated coupons with optional filters.
func (s *couponService) List(filters dto.CouponListFilters) ([]models.Coupon, int64, error) {
	var coupons []models.Coupon
	var total int64
	now := time.Now()

	query := s.db.Model(&models.Coupon{})

	// 1. Mandatory usability filters (Hide exhausted and expired coupons)
	// Only show coupons that have usage left AND are within the valid date range
	query = query.Where("used_count < usage_limit AND start_date <= ? AND end_date >= ?", now, now)

	// 2. Optional user-provided filters
	if filters.Code != "" {
		query = query.Where("code LIKE ?", "%"+filters.Code+"%")
	}

	// Handle IsActive: Default to true if not provided, otherwise use user input
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

	// 3. Count total (must be done after all filters are applied)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	// 4. Paginate
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filters.Offset

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return coupons, total, nil
}

// RecordUsage is the core function to atomically increment used count and create usage.
// It can be called inside an existing transaction (if tx != nil) or creates its own.
func (s *couponService) RecordUsage(tx *gorm.DB, couponID, userID, orderID uint, discountAmount float64) error {
	// Use provided transaction or start a new one
	exec := s.db
	if tx != nil {
		exec = tx
	} else {
		exec = exec.Begin()
		defer func() {
			if r := recover(); r != nil {
				exec.Rollback()
			}
		}()
	}

	// Increment used_count
	if err := exec.Model(&models.Coupon{}).Where("id = ?", couponID).
		Update("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
		if tx == nil {
			exec.Rollback()
		}
		return utils.ErrInternal(err)
	}

	// Create usage record
	usage := &models.CouponUsage{
		CouponID:       couponID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discountAmount,
	}
	if err := exec.Create(usage).Error; err != nil {
		if tx == nil {
			exec.Rollback()
		}
		return utils.ErrInternal(err)
	}

	if tx == nil {
		if err := exec.Commit().Error; err != nil {
			return utils.ErrInternal(err)
		}
	}
	return nil
}

// GetAvailableCouponsForUser returns coupons that are valid and not yet used by the user
func (s *couponService) GetAvailableCouponsForUser(userID uint, orderTotal float64) ([]models.Coupon, error) {
	var coupons []models.Coupon
	now := time.Now()

	// Get all active, valid, and not exhausted coupons
	query := s.db.Model(&models.Coupon{}).
		Where("is_active = ? AND used_count < usage_limit AND start_date <= ? AND end_date >= ?",
			true, now, now)

	// Exclude coupons already used by this user
	query = query.Where("id NOT IN (?)",
		s.db.Model(&models.CouponUsage{}).
			Select("coupon_id").
			Where("user_id = ?", userID))

	// Only show coupons that meet the minimum order amount
	if orderTotal > 0 {
		query = query.Where("minimum_order_amount <= ?", orderTotal)
	}

	if err := query.Order("discount_value DESC").Find(&coupons).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return coupons, nil
}
