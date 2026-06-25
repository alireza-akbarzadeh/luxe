package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PaymentRepository implements payment persistence with GORM.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a GORM-backed payment repository.
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// CreatePaymentTx inserts a payment inside an existing transaction.
func (r *PaymentRepository) CreatePaymentTx(tx *gorm.DB, payment *models.Payment) error {
	return tx.Create(payment).Error
}

// FindPaymentTx loads a payment inside a transaction.
func (r *PaymentRepository) FindPaymentTx(tx *gorm.DB, paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := tx.First(&payment, paymentID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// SavePaymentTx persists payment changes inside a transaction.
func (r *PaymentRepository) SavePaymentTx(tx *gorm.DB, payment *models.Payment) error {
	return tx.Save(payment).Error
}

// UpdatePaymentFields updates selected columns on a payment.
func (r *PaymentRepository) UpdatePaymentFields(ctx context.Context, payment *models.Payment, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(payment).Updates(updates).Error
}

// FindByStripeSession loads a payment by Stripe session ID.
func (r *PaymentRepository) FindByStripeSession(ctx context.Context, sessionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).Where("stripe_session_id = ?", sessionID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdatePaymentByID updates payment columns by ID.
func (r *PaymentRepository) UpdatePaymentByID(ctx context.Context, paymentID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", paymentID).Updates(updates).Error
}

// ListActiveProviders returns active payment providers.
func (r *PaymentRepository) ListActiveProviders(ctx context.Context, isActive bool) ([]models.PaymentProviders, error) {
	var methods []models.PaymentProviders
	err := r.db.WithContext(ctx).Where("is_active = ?", isActive).
		Order("sort_order ASC").
		Find(&methods).Error
	return methods, err
}

// DB returns the underlying GORM handle.
func (r *PaymentRepository) DB() *gorm.DB {
	return r.db
}
