package postgres

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// GiftCardRepository persists gift cards with GORM.
type GiftCardRepository struct {
	db *gorm.DB
}

// NewGiftCardRepository creates a GORM-backed gift card repository.
func NewGiftCardRepository(db *gorm.DB) *GiftCardRepository {
	return &GiftCardRepository{db: db}
}

// Create inserts a gift card row.
func (r *GiftCardRepository) Create(ctx context.Context, card *models.GiftCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

// FindByCode loads a gift card by code.
func (r *GiftCardRepository) FindByCode(ctx context.Context, code string) (*models.GiftCard, error) {
	var card models.GiftCard
	err := r.db.WithContext(ctx).Where("code = ?", strings.ToUpper(strings.TrimSpace(code))).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

// CountSent counts gift cards sent by a user.
func (r *GiftCardRepository) CountSent(ctx context.Context, senderUserID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.GiftCard{}).Where("sender_user_id = ?", senderUserID).Count(&total).Error
	return total, err
}

// ListSent returns paginated gift cards sent by a user.
func (r *GiftCardRepository) ListSent(ctx context.Context, senderUserID uint, limit, offset int) ([]models.GiftCard, error) {
	var cards []models.GiftCard
	err := r.db.WithContext(ctx).
		Where("sender_user_id = ?", senderUserID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&cards).Error
	return cards, err
}

// CountReceived counts gift cards received by a user (by user id or email).
func (r *GiftCardRepository) CountReceived(ctx context.Context, userID uint, email string) (int64, error) {
	var total int64
	q := r.db.WithContext(ctx).Model(&models.GiftCard{})
	if userID > 0 && email != "" {
		q = q.Where("recipient_user_id = ? OR LOWER(recipient_email) = LOWER(?)", userID, email)
	} else if userID > 0 {
		q = q.Where("recipient_user_id = ?", userID)
	} else if email != "" {
		q = q.Where("LOWER(recipient_email) = LOWER(?)", email)
	}
	err := q.Count(&total).Error
	return total, err
}

// ListReceived returns paginated gift cards received by a user.
func (r *GiftCardRepository) ListReceived(ctx context.Context, userID uint, email string, limit, offset int) ([]models.GiftCard, error) {
	var cards []models.GiftCard
	q := r.db.WithContext(ctx)
	if userID > 0 && email != "" {
		q = q.Where("recipient_user_id = ? OR LOWER(recipient_email) = LOWER(?)", userID, email)
	} else if userID > 0 {
		q = q.Where("recipient_user_id = ?", userID)
	} else if email != "" {
		q = q.Where("LOWER(recipient_email) = LOWER(?)", email)
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&cards).Error
	return cards, err
}

// Save persists gift card changes.
func (r *GiftCardRepository) Save(ctx context.Context, card *models.GiftCard) error {
	return r.db.WithContext(ctx).Save(card).Error
}

// CodeExists reports whether a gift card code is already taken.
func (r *GiftCardRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.GiftCard{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}
