package giftcard

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

const giftCardStripeMetadataType = "gift_card_purchase"

// Service orchestrates gift card use cases.
type Service struct {
	repo          *postgres.GiftCardRepository
	stripe        *stripeintegration.Gateway
	stripeEnabled bool
}

// NewService creates gift card use cases.
func NewService(
	repo *postgres.GiftCardRepository,
	stripe *stripeintegration.Gateway,
	stripeEnabled bool,
) *Service {
	return &Service{
		repo:          repo,
		stripe:        stripe,
		stripeEnabled: stripeEnabled,
	}
}

func generateGiftCardCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	segment := func(length int) (string, error) {
		out := make([]byte, length)
		for i := range out {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
			if err != nil {
				return "", err
			}
			out[i] = alphabet[n.Int64()]
		}
		return string(out), nil
	}
	a, err := segment(4)
	if err != nil {
		return "", err
	}
	b, err := segment(4)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("LUXE-%s-%s", a, b), nil
}

func parseDeliveryDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, errors.New("empty delivery date")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			normalized := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
			return &normalized, nil
		}
	}

	return nil, errors.New("invalid delivery date")
}

func (s *Service) uniqueCode(ctx context.Context) (string, error) {
	for i := 0; i < 8; i++ {
		code, err := generateGiftCardCode()
		if err != nil {
			return "", err
		}
		exists, err := s.repo.CodeExists(ctx, code)
		if err != nil {
			return "", utils.ErrInternal(err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", utils.ErrInternal(errors.New("failed to generate unique gift card code"))
}

// Create initiates a gift card purchase. When Stripe is enabled, returns a checkout URL;
// otherwise activates the card immediately (local/dev).
func (s *Service) Create(
	ctx context.Context,
	senderUserID uint,
	customerEmail string,
	req dto.CreateGiftCardRequest,
) (*dto.CreateGiftCardResponse, error) {
	code, err := s.uniqueCode(ctx)
	if err != nil {
		return nil, err
	}

	var deliveryDate *time.Time
	if strings.TrimSpace(req.DeliveryDate) != "" {
		parsed, parseErr := parseDeliveryDate(req.DeliveryDate)
		if parseErr != nil {
			return nil, utils.ErrBadRequest("invalid delivery_date format, use YYYY-MM-DD")
		}
		deliveryDate = parsed
	}

	expiresAt := time.Now().AddDate(1, 0, 0)
	card := &models.GiftCard{
		Code:           code,
		SenderUserID:   senderUserID,
		RecipientEmail: strings.ToLower(strings.TrimSpace(req.RecipientEmail)),
		RecipientName:  strings.TrimSpace(req.RecipientName),
		SenderName:     strings.TrimSpace(req.SenderName),
		Message:        strings.TrimSpace(req.Message),
		InitialAmount:  req.Amount,
		Balance:        0,
		Currency:       "USD",
		Status:         constants.GiftCardStatusPending,
		DeliveryDate:   deliveryDate,
		ExpiresAt:      &expiresAt,
	}

	if err := s.repo.Create(ctx, card); err != nil {
		return nil, utils.ErrInternal(err)
	}

	response := &dto.CreateGiftCardResponse{
		GiftCardResponse: dto.ToGiftCardResponse(card),
	}

	if !s.stripeEnabled || s.stripe == nil {
		_ = s.cancelCard(ctx, card)
		return nil, utils.ErrBadRequest("card payment is not available — configure Stripe to purchase gift cards")
	}

	if strings.TrimSpace(customerEmail) == "" {
		_ = s.cancelCard(ctx, card)
		return nil, utils.ErrBadRequest("account email is required for card payment")
	}

	checkoutURL, sessionID, err := s.stripe.CreateGiftCardSession(
		card.ID,
		req.Amount,
		card.Currency,
		customerEmail,
		card.RecipientName,
	)
	if err != nil {
		_ = s.cancelCard(ctx, card)
		return nil, utils.ErrInternal(err)
	}

	card.StripeSessionID = sessionID
	if err := s.repo.Save(ctx, card); err != nil {
		_ = s.cancelCard(ctx, card)
		return nil, utils.ErrInternal(err)
	}

	response.CheckoutURL = checkoutURL
	response.GiftCardResponse = dto.ToGiftCardResponse(card)
	return response, nil
}

func (s *Service) activateCard(ctx context.Context, card *models.GiftCard) error {
	card.Status = constants.GiftCardStatusActive
	card.Balance = card.InitialAmount
	if err := s.repo.Save(ctx, card); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (s *Service) cancelCard(ctx context.Context, card *models.GiftCard) error {
	card.Status = constants.GiftCardStatusCancelled
	return s.repo.Save(ctx, card)
}

// ConfirmByStripeSession activates a gift card after successful Stripe payment.
func (s *Service) ConfirmByStripeSession(ctx context.Context, sessionID, paymentIntentID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return utils.ErrBadRequest("session id is required")
	}

	card, err := s.repo.FindByStripeSessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("gift card not found for session")
		}
		return utils.ErrInternal(err)
	}

	if card.Status == constants.GiftCardStatusActive {
		return nil
	}
	if card.Status != constants.GiftCardStatusPending {
		return utils.ErrBadRequest("gift card is not awaiting payment")
	}

	_ = paymentIntentID
	return s.activateCard(ctx, card)
}

// ConfirmBySessionID confirms payment using the session ID from the Stripe success redirect.
func (s *Service) ConfirmBySessionID(ctx context.Context, senderUserID uint, sessionID string) (dto.GiftCardResponse, error) {
	if !s.stripeEnabled || s.stripe == nil {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("card payment is not configured")
	}

	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("session_id is required")
	}

	checkoutSession, err := s.stripe.GetCheckoutSession(sessionID)
	if err != nil {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("invalid or expired checkout session")
	}

	if checkoutSession.Metadata == nil || checkoutSession.Metadata["type"] != giftCardStripeMetadataType {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("not a gift card checkout session")
	}

	cardIDStr := checkoutSession.Metadata["gift_card_id"]
	cardID64, parseErr := strconv.ParseUint(cardIDStr, 10, 64)
	if parseErr != nil || cardID64 == 0 {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("invalid gift card reference")
	}

	card, err := s.repo.FindByStripeSessionID(ctx, checkoutSession.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.GiftCardResponse{}, utils.ErrNotFound("gift card not found")
		}
		return dto.GiftCardResponse{}, utils.ErrInternal(err)
	}

	if card.ID != uint(cardID64) || card.SenderUserID != senderUserID {
		return dto.GiftCardResponse{}, utils.ErrForbidden("checkout session does not belong to this account")
	}

	if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("payment is not completed yet — refresh in a moment")
	}

	paymentIntentID := ""
	if checkoutSession.PaymentIntent != nil {
		paymentIntentID = checkoutSession.PaymentIntent.ID
	}

	if err := s.ConfirmByStripeSession(ctx, checkoutSession.ID, paymentIntentID); err != nil {
		return dto.GiftCardResponse{}, err
	}

	updated, err := s.repo.FindByStripeSessionID(ctx, checkoutSession.ID)
	if err != nil {
		return dto.GiftCardResponse{}, utils.ErrInternal(err)
	}

	return dto.ToGiftCardResponse(updated), nil
}

// FailByStripeSession cancels a pending gift card when checkout expires.
func (s *Service) FailByStripeSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}

	card, err := s.repo.FindByStripeSessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return utils.ErrInternal(err)
	}

	if card.Status != constants.GiftCardStatusPending {
		return nil
	}

	return s.cancelCard(ctx, card)
}

// ListSent returns gift cards created by the user.
func (s *Service) ListSent(ctx context.Context, senderUserID uint, limit, offset int) ([]models.GiftCard, int64, error) {
	total, err := s.repo.CountSent(ctx, senderUserID)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	cards, err := s.repo.ListSent(ctx, senderUserID, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return cards, total, nil
}

// ListReceived returns gift cards addressed to the user.
func (s *Service) ListReceived(ctx context.Context, userID uint, email string, limit, offset int) ([]models.GiftCard, int64, error) {
	total, err := s.repo.CountReceived(ctx, userID, email)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	cards, err := s.repo.ListReceived(ctx, userID, email, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	for i := range cards {
		if cards[i].RecipientUserID == nil && strings.EqualFold(cards[i].RecipientEmail, email) {
			uid := userID
			cards[i].RecipientUserID = &uid
		}
	}
	return cards, total, nil
}

// Claim associates a gift card with the authenticated recipient by code.
func (s *Service) Claim(ctx context.Context, userID uint, email, code string) (*models.GiftCard, error) {
	card, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("gift card not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if !strings.EqualFold(card.RecipientEmail, email) {
		return nil, utils.ErrForbidden("this gift card is not addressed to your account")
	}
	if card.Status != constants.GiftCardStatusActive {
		return nil, utils.ErrBadRequest("gift card is not active")
	}
	if card.RecipientUserID != nil && *card.RecipientUserID != userID {
		return nil, utils.ErrForbidden("gift card already claimed by another account")
	}
	if card.RecipientUserID == nil {
		card.RecipientUserID = &userID
		if err := s.repo.Save(ctx, card); err != nil {
			return nil, utils.ErrInternal(err)
		}
	}
	return card, nil
}

// SpendForMembership deducts from a claimed gift card to pay for Luxe Plus.
func (s *Service) SpendForMembership(ctx context.Context, userID uint, email, code string, amount float64) error {
	card, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("gift card not found")
		}
		return utils.ErrInternal(err)
	}
	if !strings.EqualFold(card.RecipientEmail, email) {
		return utils.ErrForbidden("this gift card is not addressed to your account")
	}
	if card.Status != constants.GiftCardStatusActive {
		return utils.ErrBadRequest("gift card is not active")
	}
	if card.ExpiresAt != nil && card.ExpiresAt.Before(time.Now()) {
		return utils.ErrBadRequest("gift card has expired")
	}
	if card.Balance < amount {
		return utils.ErrBadRequest("insufficient gift card balance")
	}
	if card.RecipientUserID != nil && *card.RecipientUserID != userID {
		return utils.ErrForbidden("gift card already claimed by another account")
	}

	card.Balance -= amount
	if card.Balance <= 0 {
		card.Balance = 0
		now := time.Now()
		card.Status = constants.GiftCardStatusRedeemed
		card.RedeemedAt = &now
	}
	if card.RecipientUserID == nil {
		card.RecipientUserID = &userID
	}

	if err := s.repo.Save(ctx, card); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}
