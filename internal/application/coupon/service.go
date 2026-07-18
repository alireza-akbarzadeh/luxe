package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domaincoupon "github.com/alireza-akbarzadeh/luxe/internal/domain/coupon"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service orchestrates coupon use cases.
type Service struct {
	repo   *postgres.CouponRepository
	engine *workflow.Engine
}

// NewService creates coupon use cases.
func NewService(repo *postgres.CouponRepository, engine *workflow.Engine) *Service {
	return &Service{repo: repo, engine: engine}
}

func (s *Service) syncCouponWorkflow(ctx context.Context, couponID uint, isActive bool) {
	if !appworkflow.ApplyCouponWorkflow(ctx, s.engine, couponID, isActive, nil) {
		utils.Log.WithField("coupon_id", couponID).Debug("coupon workflow sync skipped or failed")
	}
}

// Create stores a new coupon.
func (s *Service) Create(ctx context.Context, req dto.CreateCouponRequest) (*models.Coupon, error) {
	appType := normalizeApplicationType(req.ApplicationType)
	code := resolveCouponCode(req.Code, appType)

	exists, err := s.repo.ExistsByCode(ctx, code, 0)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if exists {
		return nil, utils.ErrConflict("coupon code already exists")
	}

	bogoBuy := req.BogoBuyQuantity
	if bogoBuy <= 0 {
		bogoBuy = 1
	}
	bogoGet := req.BogoGetQuantity
	if bogoGet <= 0 {
		bogoGet = 1
	}
	bogoPercent := req.BogoGetDiscountPercent
	if bogoPercent <= 0 {
		bogoPercent = 100
	}

	coupon := &models.Coupon{
		Code:                   code,
		Description:            req.Description,
		ApplicationType:        appType,
		DiscountType:           req.DiscountType,
		DiscountValue:          req.DiscountValue,
		MinimumOrderAmount:     req.MinimumOrderAmount,
		MaxDiscountAmount:      req.MaxDiscountAmount,
		UsageLimit:             req.UsageLimit,
		StartDate:              normalizeCouponStart(req.StartDate),
		EndDate:                normalizeCouponEnd(req.StartDate, req.EndDate),
		IsActive:               couponIsActiveDefault(req.IsActive),
		Conditions:             models.MarshalCouponConditions(conditionsFromRequest(req.Conditions)),
		BogoBuyQuantity:        bogoBuy,
		BogoGetQuantity:        bogoGet,
		BogoGetDiscountPercent: bogoPercent,
	}
	if err := s.repo.CreateCoupon(ctx, coupon); err != nil {
		return nil, utils.ErrInternal(err)
	}
	if !coupon.IsActive {
		if err := s.repo.UpdateCouponIsActive(ctx, coupon.ID, false); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}

	if coupon.IsActive {
		s.syncCouponWorkflow(ctx, coupon.ID, true)
	} else {
		appworkflow.SyncState(ctx, s.engine, constants.WorkflowEntityCoupon, coupon.ID, "draft", "created", nil)
	}

	return s.repo.FindByID(ctx, coupon.ID)
}

func (s *Service) ValidateCoupon(ctx context.Context, code string, userID uint, orderTotal float64, itemCount int) (*models.Coupon, float64, error) {
	couponModel, err := s.repo.FindActiveByCode(ctx, code, time.Now())
	if err != nil {
		return nil, 0, utils.ErrBadRequest("invalid or expired coupon")
	}

	usageCount, err := s.repo.CountCouponUsageByUser(ctx, couponModel.ID, userID)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	conditions := models.ParseCouponConditions(couponModel.Conditions)
	if !models.CouponConditionsAllowsUser(conditions, userID) {
		return nil, 0, utils.ErrBadRequest("this promotion is not available for your account")
	}

	domainCoupon := couponFromModel(*couponModel)
	if err := domaincoupon.ValidateEligibility(domainCoupon, orderTotal, int(usageCount), itemCount, time.Now()); err != nil {
		switch {
		case errors.Is(err, domaincoupon.ErrInactiveOrExpired):
			return nil, 0, utils.ErrBadRequest("invalid or expired coupon")
		case errors.Is(err, domaincoupon.ErrUsageLimitExceeded):
			return nil, 0, utils.ErrBadRequest("coupon usage limit exceeded")
		case errors.Is(err, domaincoupon.ErrBelowMinimumOrder):
			return nil, 0, utils.ErrBadRequest("order total below minimum amount")
		case errors.Is(err, domaincoupon.ErrAlreadyUsedByUser):
			return nil, 0, utils.ErrBadRequest("coupon already used by this user")
		case errors.Is(err, domaincoupon.ErrInsufficientItems):
			return nil, 0, utils.ErrBadRequest("cart does not meet promotion item requirements")
		default:
			return nil, 0, utils.ErrInternal(err)
		}
	}

	discount := domaincoupon.CalculateDiscount(domainCoupon, orderTotal, itemCount)
	return couponModel, discount, nil
}

func (s *Service) ApplyCoupon(tx *gorm.DB, userID uint, orderID uint, couponCode string, orderTotal float64, itemCount int) error {
	coupon, discount, err := s.ValidateCoupon(context.Background(), couponCode, userID, orderTotal, itemCount)
	if err != nil {
		return err
	}

	if err := s.repo.IncrementUsedCountTx(tx, coupon); err != nil {
		return utils.ErrInternal(err)
	}

	if coupon.UsageLimit > 0 && domaincoupon.IsExhaustedAfterUse(couponFromModel(*coupon)) {
		appworkflow.ApplyCouponExhausted(context.Background(), s.engine, coupon.ID)
	}

	usage := &models.CouponUsage{
		CouponID:       coupon.ID,
		UserID:         userID,
		OrderID:        orderID,
		DiscountAmount: discount,
	}
	if err := s.repo.CreateCouponUsageTx(tx, usage); err != nil {
		return utils.ErrInternal(err)
	}

	if err := s.repo.ApplyOrderDiscountTx(tx, orderID, discount); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *Service) GetByID(ctx context.Context, couponID uint) (*models.Coupon, error) {
	coupon, err := s.repo.FindByID(ctx, couponID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("coupon not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return coupon, nil
}

func (s *Service) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	coupon, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("coupon not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return coupon, nil
}

func (s *Service) Update(ctx context.Context, id uint, req dto.UpdateCouponRequest) (*models.Coupon, error) {
	coupon, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Code != nil && *req.Code != "" {
		exists, err := s.repo.ExistsByCode(ctx, *req.Code, id)
		if err != nil {
			return nil, utils.ErrInternal(err)
		}
		if exists {
			return nil, utils.ErrConflict("coupon code already exists")
		}
		coupon.Code = strings.TrimSpace(strings.ToUpper(*req.Code))
	}
	if req.Description != nil {
		coupon.Description = *req.Description
	}
	if req.ApplicationType != nil {
		coupon.ApplicationType = normalizeApplicationType(*req.ApplicationType)
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
	wasActive := coupon.IsActive
	if req.IsActive != nil {
		coupon.IsActive = *req.IsActive
	}
	if req.Conditions != nil {
		applyConditionsRequest(coupon, req.Conditions)
	}
	if req.BogoBuyQuantity != nil && *req.BogoBuyQuantity > 0 {
		coupon.BogoBuyQuantity = *req.BogoBuyQuantity
	}
	if req.BogoGetQuantity != nil && *req.BogoGetQuantity > 0 {
		coupon.BogoGetQuantity = *req.BogoGetQuantity
	}
	if req.BogoGetDiscountPercent != nil {
		coupon.BogoGetDiscountPercent = *req.BogoGetDiscountPercent
	}

	if err := s.repo.SaveCoupon(ctx, coupon); err != nil {
		return nil, utils.ErrInternal(err)
	}
	if req.IsActive != nil && wasActive != *req.IsActive {
		s.syncCouponWorkflow(ctx, coupon.ID, *req.IsActive)
	}
	return s.repo.FindByID(ctx, coupon.ID)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	rows, err := s.repo.DeleteCoupon(ctx, id)
	if err != nil {
		return utils.ErrInternal(err)
	}
	if rows == 0 {
		return utils.ErrNotFound("coupon not found")
	}
	return nil
}

func (s *Service) List(ctx context.Context, filters dto.CouponListFilters) ([]models.Coupon, int64, error) {
	coupons, total, err := s.repo.ListPublic(ctx, filters, time.Now())
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return coupons, total, nil
}

func (s *Service) ListAdmin(ctx context.Context, filters dto.AdminCouponListFilters) ([]models.Coupon, int64, error) {
	coupons, total, err := s.repo.ListAdmin(ctx, filters, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			return nil, 0, utils.ErrBadRequest("invalid status filter")
		}
		return nil, 0, utils.ErrInternal(err)
	}
	return coupons, total, nil
}

func (s *Service) RecordUsage(tx *gorm.DB, couponID, userID, orderID uint, discountAmount float64) error {
	exec := s.repo.DB()
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

	if err := s.repo.RecordUsageTx(exec, couponID, userID, orderID, discountAmount); err != nil {
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

func (s *Service) GetAvailableCouponsForUser(ctx context.Context, userID uint, orderTotal float64) ([]models.Coupon, error) {
	coupons, err := s.repo.ListAvailableForUser(ctx, userID, orderTotal, time.Now())
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return filterCouponsForUser(coupons, userID), nil
}

func filterCouponsForUser(coupons []models.Coupon, userID uint) []models.Coupon {
	if len(coupons) == 0 {
		return coupons
	}
	filtered := make([]models.Coupon, 0, len(coupons))
	for _, coupon := range coupons {
		conditions := models.ParseCouponConditions(coupon.Conditions)
		if models.CouponConditionsAllowsUser(conditions, userID) {
			filtered = append(filtered, coupon)
		}
	}
	return filtered
}

func normalizeCouponStart(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func normalizeCouponEnd(start, end time.Time) time.Time {
	if !end.IsZero() {
		return end
	}
	base := start
	if base.IsZero() {
		base = time.Now()
	}
	return base.AddDate(1, 0, 0)
}

func couponIsActiveDefault(isActive *bool) bool {
	if isActive == nil {
		return false
	}
	return *isActive
}

// BestAutomaticCoupon returns the highest-value automatic promotion for a cart.
func (s *Service) BestAutomaticCoupon(ctx context.Context, userID uint, orderTotal float64, itemCount int) (*models.Coupon, float64, error) {
	coupons, err := s.repo.ListAvailableForUser(ctx, userID, orderTotal, time.Now())
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	var best *models.Coupon
	var bestDiscount float64

	for i := range coupons {
		if normalizeApplicationType(coupons[i].ApplicationType) != constants.CouponApplicationAutomatic {
			continue
		}
		_, discount, err := s.ValidateCoupon(ctx, coupons[i].Code, userID, orderTotal, itemCount)
		if err != nil {
			continue
		}
		if discount > bestDiscount {
			copy := coupons[i]
			best = &copy
			bestDiscount = discount
		}
	}

	if best == nil {
		return nil, 0, utils.ErrNotFound("no automatic promotion available")
	}
	return best, bestDiscount, nil
}
