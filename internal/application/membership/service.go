package membership

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	appgiftcard "github.com/alireza-akbarzadeh/luxe/internal/application/giftcard"
	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

// Service orchestrates Luxe Plus membership use cases.
type Service struct {
	users         *postgres.UserRepository
	wallet        *appwallet.Service
	giftCards     *appgiftcard.Service
	stripe        *stripeintegration.Gateway
	stripeEnabled bool
}

// NewService wires membership dependencies.
func NewService(
	users *postgres.UserRepository,
	wallet *appwallet.Service,
	giftCards *appgiftcard.Service,
	stripe *stripeintegration.Gateway,
	stripeEnabled bool,
) *Service {
	return &Service{
		users:         users,
		wallet:        wallet,
		giftCards:     giftCards,
		stripe:        stripe,
		stripeEnabled: stripeEnabled,
	}
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

// Subscribe activates Luxe Plus using wallet, gift card, or Stripe Checkout.
func (s *Service) Subscribe(ctx context.Context, userID uint, email string, req dto.SubscribePlusRequest) (dto.SubscribePlusResponse, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.SubscribePlusResponse{}, utils.ErrNotFound("user not found")
		}
		return dto.SubscribePlusResponse{}, utils.ErrInternal(err)
	}

	if IsPlusActive(user) {
		return dto.SubscribePlusResponse{}, utils.ErrConflict("Luxe Plus is already active on your account")
	}

	price := constants.PlusAnnualPriceUSD
	method := strings.TrimSpace(req.PaymentMethod)

	switch method {
	case constants.PlusPaymentWallet:
		return s.subscribeWithWallet(ctx, user, price)
	case constants.PlusPaymentGiftCard:
		code := strings.TrimSpace(req.GiftCardCode)
		if code == "" {
			return dto.SubscribePlusResponse{}, utils.ErrBadRequest("gift_card_code is required")
		}
		return s.subscribeWithGiftCard(ctx, user, email, code, price)
	case constants.PlusPaymentStripe:
		return s.subscribeWithStripe(ctx, user, email, price)
	default:
		return dto.SubscribePlusResponse{}, utils.ErrBadRequest("invalid payment_method")
	}
}

func (s *Service) subscribeWithWallet(ctx context.Context, user *models.User, price float64) (dto.SubscribePlusResponse, error) {
	balance, err := s.wallet.GetBalance(ctx, user.ID)
	if err != nil {
		return dto.SubscribePlusResponse{}, err
	}
	if balance < price {
		return dto.SubscribePlusResponse{}, utils.ErrBadRequest("insufficient wallet balance — top up your wallet or choose another payment method")
	}

	if err := s.wallet.Withdraw(ctx, user.ID, price, "membership", nil, "Luxe Plus annual membership"); err != nil {
		return dto.SubscribePlusResponse{}, err
	}

	if err := s.activatePlus(ctx, user); err != nil {
		return dto.SubscribePlusResponse{}, err
	}

	return dto.SubscribePlusResponse{
		Membership:    ToStatusResponse(user),
		PaymentMethod: constants.PlusPaymentWallet,
		PaymentStatus: constants.PlusPaymentStatusCompleted,
	}, nil
}

func (s *Service) subscribeWithGiftCard(ctx context.Context, user *models.User, email, code string, price float64) (dto.SubscribePlusResponse, error) {
	if err := s.giftCards.SpendForMembership(ctx, user.ID, email, code, price); err != nil {
		return dto.SubscribePlusResponse{}, err
	}

	if err := s.activatePlus(ctx, user); err != nil {
		return dto.SubscribePlusResponse{}, err
	}

	return dto.SubscribePlusResponse{
		Membership:    ToStatusResponse(user),
		PaymentMethod: constants.PlusPaymentGiftCard,
		PaymentStatus: constants.PlusPaymentStatusCompleted,
	}, nil
}

func (s *Service) subscribeWithStripe(ctx context.Context, user *models.User, email string, price float64) (dto.SubscribePlusResponse, error) {
	if !s.stripeEnabled || s.stripe == nil {
		if err := s.activatePlus(ctx, user); err != nil {
			return dto.SubscribePlusResponse{}, err
		}
		return dto.SubscribePlusResponse{
			Membership:    ToStatusResponse(user),
			PaymentMethod: constants.PlusPaymentStripe,
			PaymentStatus: constants.PlusPaymentStatusCompleted,
		}, nil
	}

	if strings.TrimSpace(email) == "" {
		return dto.SubscribePlusResponse{}, utils.ErrBadRequest("customer email is required for card payment")
	}

	checkoutURL, sessionID, err := s.stripe.CreatePlusMembershipSession(user.ID, price, "USD", email)
	if err != nil {
		return dto.SubscribePlusResponse{}, utils.ErrInternal(err)
	}

	return dto.SubscribePlusResponse{
		Membership:      ToStatusResponse(user),
		PaymentMethod:   constants.PlusPaymentStripe,
		PaymentStatus:   constants.PlusPaymentStatusPending,
		CheckoutURL:     checkoutURL,
		StripeSessionID: sessionID,
	}, nil
}

// ConfirmStripeSubscription activates Plus after Stripe Checkout completes (webhook).
func (s *Service) ConfirmStripeSubscription(ctx context.Context, session stripe.CheckoutSession) error {
	if session.Metadata == nil || session.Metadata["type"] != constants.PlusStripeMetadataType {
		return utils.ErrBadRequest("invalid plus membership session")
	}
	if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return utils.ErrBadRequest("payment not completed")
	}

	userIDStr := session.Metadata["user_id"]
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID64 == 0 {
		return utils.ErrBadRequest("invalid user in session metadata")
	}

	user, err := s.users.FindByID(ctx, uint(userID64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("user not found")
		}
		return utils.ErrInternal(err)
	}

	if IsPlusActive(user) {
		return nil
	}

	expectedCents := int64(constants.PlusAnnualPriceUSD * 100)
	if session.AmountTotal != expectedCents {
		return utils.ErrBadRequest(fmt.Sprintf("unexpected payment amount: got %d want %d", session.AmountTotal, expectedCents))
	}

	return s.activatePlus(ctx, user)
}

func (s *Service) activatePlus(ctx context.Context, user *models.User) error {
	now := time.Now()
	expires := now.AddDate(1, 0, 0)
	user.MembershipTier = constants.MembershipTierPlus
	user.PlusSubscribedAt = &now
	user.PlusExpiresAt = &expires

	if err := s.users.SaveUser(ctx, user); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
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
