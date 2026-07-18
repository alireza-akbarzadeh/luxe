package giftcard

import (
	"context"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

func displayName(user *models.User) string {
	if user == nil {
		return ""
	}
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name != "" {
		return name
	}
	return user.Email
}

func maskEmail(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return ""
	}
	visible := string(parts[0][0])
	return visible + "***@" + parts[1]
}

func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) <= 4 {
		return ""
	}
	return "••••" + phone[len(phone)-4:]
}

func isGiftCardActor(card *models.GiftCard, userID uint, email string) bool {
	if card.SenderUserID == userID {
		return true
	}
	if card.RecipientUserID != nil && *card.RecipientUserID == userID {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(card.RecipientEmail), strings.TrimSpace(email))
}

// LookupRecipients finds Luxe members to receive a gift card (name, email, or phone search).
func (s *Service) LookupRecipients(ctx context.Context, actorUserID uint, query string) ([]dto.GiftRecipientLookupResponse, error) {
	users, err := s.users.SearchGiftRecipients(ctx, actorUserID, query, 10)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	items := make([]dto.GiftRecipientLookupResponse, 0, len(users))
	for i := range users {
		user := users[i]
		items = append(items, dto.GiftRecipientLookupResponse{
			ID:          user.ID,
			DisplayName: displayName(&user),
			MaskedEmail: maskEmail(user.Email),
			MaskedPhone: maskPhone(user.Phone),
		})
	}
	return items, nil
}

// Transfer reassigns an active gift card to another Luxe member.
func (s *Service) Transfer(
	ctx context.Context,
	actorUserID uint,
	actorEmail string,
	code string,
	req dto.TransferGiftCardRequest,
) (dto.GiftCardResponse, error) {
	card, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.GiftCardResponse{}, utils.ErrNotFound("gift card not found")
		}
		return dto.GiftCardResponse{}, utils.ErrInternal(err)
	}

	if !isGiftCardActor(card, actorUserID, actorEmail) {
		return dto.GiftCardResponse{}, utils.ErrForbidden("you cannot transfer this gift card")
	}
	if card.Status != constants.GiftCardStatusActive {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("only active gift cards can be transferred")
	}
	if card.Balance <= 0 {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("gift card has no remaining balance")
	}
	if req.RecipientUserID == actorUserID {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("you cannot transfer a gift card to yourself")
	}

	target, err := s.users.FindByID(ctx, req.RecipientUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.GiftCardResponse{}, utils.ErrNotFound("recipient user not found")
		}
		return dto.GiftCardResponse{}, utils.ErrInternal(err)
	}
	if !target.IsActive {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("recipient account is not active")
	}
	if strings.EqualFold(strings.TrimSpace(target.Email), strings.TrimSpace(card.RecipientEmail)) &&
		card.RecipientUserID != nil && *card.RecipientUserID == target.ID {
		return dto.GiftCardResponse{}, utils.ErrBadRequest("gift card is already assigned to this user")
	}

	previousRecipientID := card.RecipientUserID

	card.RecipientEmail = strings.ToLower(strings.TrimSpace(target.Email))
	card.RecipientName = displayName(target)
	recipientID := target.ID
	card.RecipientUserID = &recipientID

	if err := s.repo.Save(ctx, card); err != nil {
		return dto.GiftCardResponse{}, utils.ErrInternal(err)
	}

	s.notifyTransfer(card, actorUserID, previousRecipientID, target)
	return dto.ToGiftCardResponse(card), nil
}

func (s *Service) notifyTransfer(
	card *models.GiftCard,
	actorUserID uint,
	previousRecipientID *uint,
	target *models.User,
) {
	if card == nil || target == nil {
		return
	}

	amountLabel := formatMoney(card.Balance, card.Currency)
	fromName := strings.TrimSpace(card.SenderName)
	if fromName == "" {
		fromName = "A Luxe member"
	}
	if actorUserID != card.SenderUserID {
		actor, err := s.users.FindByID(context.Background(), actorUserID)
		if err == nil {
			fromName = displayName(actor)
		}
	}

	accountURL := strings.TrimRight(s.frontendURL, "/")
	if accountURL == "" {
		accountURL = "https://luxe.app/account"
	}

	data := map[string]interface{}{
		"account_tab":  "giftCards",
		"gift_card_id": card.ID,
		"code":         card.Code,
	}

	go func() {
		if s.notifier != nil {
			_ = s.notifier.CreateNotification(
				card.SenderUserID,
				constants.NotificationTypeGiftCardTransferred,
				"Gift card transferred",
				fmt.Sprintf("Your gift card (%s) was sent to %s.", amountLabel, displayName(target)),
				data,
			)
			_ = s.notifier.CreateNotification(
				target.ID,
				constants.NotificationTypeGiftCardReceived,
				"You received a gift card",
				fmt.Sprintf("%s sent you a %s Luxe gift card.", fromName, amountLabel),
				data,
			)
			if previousRecipientID != nil &&
				*previousRecipientID != target.ID &&
				*previousRecipientID != actorUserID {
				_ = s.notifier.CreateNotification(
					*previousRecipientID,
					constants.NotificationTypeGiftCardTransferred,
					"Gift card reassigned",
					"A gift card previously sent to you was reassigned to another member.",
					data,
				)
			}
		}

		if s.jobQueue != nil {
			subject, body := giftCardTransferredRecipientEmail(fromName, displayName(target), amountLabel, accountURL)
			_ = s.jobQueue.EnqueueSendEmail(context.Background(), target.Email, subject, body)

			if card.SenderUserID != actorUserID {
				sender, err := s.users.FindByID(context.Background(), card.SenderUserID)
				if err == nil && sender.Email != "" {
					sentSubject, sentBody := giftCardSentEmail(displayName(sender), displayName(target), amountLabel, accountURL)
					_ = s.jobQueue.EnqueueSendEmail(context.Background(), sender.Email, sentSubject, sentBody)
				}
			}
		}
	}()
}
