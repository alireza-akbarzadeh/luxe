package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentServiceInterface interface {
	CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error)
	ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error
}

type paymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) PaymentServiceInterface {
	return &paymentService{db: db}
}

func (s *paymentService) CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error) {
	payment := &models.Payment{
		OrderID:  req.OrderID,
		UserID:   req.UserID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Method:   req.Method,
		Status:   "pending",
	}
	if err := tx.Create(payment).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return payment, nil
}

// ProcessPayment validates the card, updates the payment record, and returns nil on success.
// On failure it sets the appropriate status and returns an error.
func (s *paymentService) ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error {
	var payment models.Payment
	if err := tx.First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	// --- MOCK GATEWAY ---
	// Simulate a real payment gateway call with test cards.
	if err := s.mockGateway(cardInfo); err != nil {
		// Payment failed – update status and gateway response
		payment.Status = "failed"
		payment.GatewayResponse = datatypes.JSON(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		if err := tx.Save(&payment).Error; err != nil {
			return fmt.Errorf("failed to save failed payment: %w", err)
		}
		return err
	}

	// Payment succeeded
	payment.Status = "succeeded"
	payment.TransactionID = fmt.Sprintf("txn_%d", time.Now().UnixNano())
	if err := tx.Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to save successful payment: %w", err)
	}
	return nil
}

// mockGateway simulates a third‑party payment processor.
// Replace this entire function when integrating Stripe / PayPal.
func (s *paymentService) mockGateway(cardInfo dto.CardInfo) error {
	// Simulate network latency (optional)
	time.Sleep(1 * time.Second)

	// Test card that always declines
	if cardInfo.CardNumber == "0000000000000000" {
		return errors.New("card declined")
	}

	// Expiry validation
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if cardInfo.ExpiryYear < year || (cardInfo.ExpiryYear == year && cardInfo.ExpiryMonth < month) {
		return errors.New("card expired")
	}

	// CVV check (just length)
	if len(cardInfo.CVV) != 3 {
		return errors.New("invalid CVV")
	}

	// All good
	return nil
}
