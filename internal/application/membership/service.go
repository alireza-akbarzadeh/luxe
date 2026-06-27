package membership

import (
	"context"
	"errors"
	"time"

	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service orchestrates Luxe Plus membership use cases.
type Service struct {
	users  *postgres.UserRepository
	wallet *appwallet.Service
}

// NewService wires membership dependencies.
func NewService(users *postgres.UserRepository, wallet *appwallet.Service) *Service {
	return &Service{users: users, wallet: wallet}
}

// BenefitsCatalog returns the public Plus plan benefits for marketing pages.
func (s *Service) BenefitsCatalog() dto.PlusBenefitsResponse {
	return dto.PlusBenefitsResponse{
		PlanName:        "Luxe Plus",
		AnnualPrice:     constants.PlusAnnualPriceUSD,
		Currency:        "USD",
		DiscountPercent: constants.PlusCheckoutDiscountPct,
		ReturnWindowDays: dto.PlusTierWindow{
			Free: constants.FreeReturnWindowDays,
			Plus: constants.PlusReturnWindowDays,
		},
		PriorityShipping: true,
		PrioritySupport:  true,
		Features: []dto.PlusBenefitFeature{
			{
				Key:         "discount",
				Title:       "Member savings",
				Description: "10% off every order, automatically applied at checkout.",
			},
			{
				Key:         "shipping",
				Title:       "Fast delivery",
				Description: "Priority processing and express shipping on eligible orders.",
			},
			{
				Key:         "returns",
				Title:       "Extended returns",
				Description: "60-day return window on delivered orders (30 days on Free).",
			},
			{
				Key:         "support",
				Title:       "Priority support",
				Description: "Skip the queue — dedicated Plus support with faster response times.",
			},
		},
	}
}

// Status returns the authenticated user's membership state.
func (s *Service) Status(ctx context.Context, userID uint) (dto.MembershipStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.MembershipStatusResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.MembershipStatusResponse{}, utils.ErrInternal(err)
	}
	return ToStatusResponse(user), nil
}

// Subscribe activates Luxe Plus for one year, charging the user's wallet.
func (s *Service) Subscribe(ctx context.Context, userID uint) (dto.MembershipStatusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.MembershipStatusResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.MembershipStatusResponse{}, utils.ErrInternal(err)
	}

	if IsPlusActive(user) {
		return dto.MembershipStatusResponse{}, utils.ErrConflict("Luxe Plus is already active on your account")
	}

	price := constants.PlusAnnualPriceUSD
	balance, err := s.wallet.GetBalance(ctx, userID)
	if err != nil {
		return dto.MembershipStatusResponse{}, err
	}
	if balance < price {
		return dto.MembershipStatusResponse{}, utils.ErrBadRequest("insufficient wallet balance — top up your wallet to subscribe")
	}

	if err := s.wallet.Withdraw(ctx, userID, price, "membership", nil, "Luxe Plus annual membership"); err != nil {
		return dto.MembershipStatusResponse{}, err
	}

	now := time.Now()
	expires := now.AddDate(1, 0, 0)
	user.MembershipTier = constants.MembershipTierPlus
	user.PlusSubscribedAt = &now
	user.PlusExpiresAt = &expires

	if err := s.users.SaveUser(ctx, user); err != nil {
		return dto.MembershipStatusResponse{}, utils.ErrInternal(err)
	}

	return ToStatusResponse(user), nil
}

// PlusOrderDiscount returns the checkout discount for an active Plus member.
func (s *Service) PlusOrderDiscount(ctx context.Context, userID uint, subtotalAfterCoupon float64) (float64, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, utils.ErrInternal(err)
	}
	return OrderDiscountAmount(user, subtotalAfterCoupon), nil
}

// UserReturnWindowDays loads the user and returns their return window in days.
func (s *Service) UserReturnWindowDays(ctx context.Context, userID uint) (int, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constants.FreeReturnWindowDays, nil
		}
		return 0, utils.ErrInternal(err)
	}
	return ReturnWindowDays(user), nil
}

// ToStatusResponse maps a user model to membership status DTO.
func ToStatusResponse(user *models.User) dto.MembershipStatusResponse {
	active := IsPlusActive(user)
	tier := constants.MembershipTierFree
	if active {
		tier = constants.MembershipTierPlus
	}
	return dto.MembershipStatusResponse{
		Tier:             tier,
		IsPlusActive:     active,
		PlusSubscribedAt: user.PlusSubscribedAt,
		PlusExpiresAt:    user.PlusExpiresAt,
		AnnualPrice:      constants.PlusAnnualPriceUSD,
		Currency:         "USD",
		Benefits:         benefitSummary(active),
	}
}

func benefitSummary(active bool) dto.PlusMemberBenefitsSummary {
	if !active {
		return dto.PlusMemberBenefitsSummary{
			DiscountPercent:  0,
			ReturnWindowDays: constants.FreeReturnWindowDays,
			PriorityShipping: false,
			PrioritySupport:  false,
		}
	}
	return dto.PlusMemberBenefitsSummary{
		DiscountPercent:  constants.PlusCheckoutDiscountPct,
		ReturnWindowDays: constants.PlusReturnWindowDays,
		PriorityShipping: true,
		PrioritySupport:  true,
	}
}
