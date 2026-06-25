package services

import (
	"context"

	appcoupon "github.com/alireza-akbarzadeh/luxe/internal/application/coupon"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type CouponServiceInterface interface {
	Create(req dto.CreateCouponRequest) (*models.Coupon, error)
	GetByID(id uint) (*models.Coupon, error)
	GetByCode(code string) (*models.Coupon, error)
	Update(id uint, req dto.UpdateCouponRequest) (*models.Coupon, error)
	Delete(id uint) error
	List(dto.CouponListFilters) ([]models.Coupon, int64, error)
	ListAdmin(dto.AdminCouponListFilters) ([]models.Coupon, int64, error)
	ValidateCoupon(code string, userID uint, orderTotal float64) (*models.Coupon, float64, error)
	ApplyCoupon(tx *gorm.DB, userID uint, orderID uint, couponCode string, orderTotal float64) error
	GetAvailableCouponsForUser(userID uint, orderTotal float64) ([]models.Coupon, error)
}

type couponService struct {
	app *appcoupon.Service
}

func NewCouponService(db *gorm.DB, engine *workflow.Engine) CouponServiceInterface {
	return &couponService{app: appcoupon.NewService(postgres.NewCouponRepository(db), engine)}
}

func (s *couponService) Create(req dto.CreateCouponRequest) (*models.Coupon, error) {
	return s.app.Create(context.Background(), req)
}

func (s *couponService) ValidateCoupon(code string, userID uint, orderTotal float64) (*models.Coupon, float64, error) {
	return s.app.ValidateCoupon(context.Background(), code, userID, orderTotal)
}

func (s *couponService) ApplyCoupon(tx *gorm.DB, userID uint, orderID uint, couponCode string, orderTotal float64) error {
	return s.app.ApplyCoupon(tx, userID, orderID, couponCode, orderTotal)
}

func (s *couponService) GetByID(couponID uint) (*models.Coupon, error) {
	return s.app.GetByID(context.Background(), couponID)
}

func (s *couponService) GetByCode(code string) (*models.Coupon, error) {
	return s.app.GetByCode(context.Background(), code)
}

func (s *couponService) Update(id uint, req dto.UpdateCouponRequest) (*models.Coupon, error) {
	return s.app.Update(context.Background(), id, req)
}

func (s *couponService) Delete(id uint) error {
	return s.app.Delete(context.Background(), id)
}

func (s *couponService) List(filters dto.CouponListFilters) ([]models.Coupon, int64, error) {
	return s.app.List(context.Background(), filters)
}

func (s *couponService) ListAdmin(filters dto.AdminCouponListFilters) ([]models.Coupon, int64, error) {
	return s.app.ListAdmin(context.Background(), filters)
}

func (s *couponService) GetAvailableCouponsForUser(userID uint, orderTotal float64) ([]models.Coupon, error) {
	return s.app.GetAvailableCouponsForUser(context.Background(), userID, orderTotal)
}
