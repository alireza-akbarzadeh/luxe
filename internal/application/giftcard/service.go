package giftcard

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Service orchestrates gift card use cases.
type Service struct {
	repo *postgres.GiftCardRepository
}

// NewService creates gift card use cases.
func NewService(repo *postgres.GiftCardRepository) *Service {
	return &Service{repo: repo}
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

// Create issues a new gift card giveaway from the authenticated user.
func (s *Service) Create(ctx context.Context, senderUserID uint, req dto.CreateGiftCardRequest) (*models.GiftCard, error) {
	code, err := s.uniqueCode(ctx)
	if err != nil {
		return nil, err
	}

	var deliveryDate *time.Time
	if strings.TrimSpace(req.DeliveryDate) != "" {
		parsed, parseErr := time.Parse("2006-01-02", req.DeliveryDate)
		if parseErr != nil {
			return nil, utils.ErrBadRequest("invalid delivery_date format, use YYYY-MM-DD")
		}
		deliveryDate = &parsed
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
		Balance:        req.Amount,
		Currency:       "USD",
		Status:         constants.GiftCardStatusActive,
		DeliveryDate:   deliveryDate,
		ExpiresAt:      &expiresAt,
	}

	if err := s.repo.Create(ctx, card); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return card, nil
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

	// Link cards to recipient when email matches but recipient_user_id is unset.
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
