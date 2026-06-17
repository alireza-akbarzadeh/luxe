package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentServiceInterface interface {
	CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error)
	ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error
	GetPaymentProvider(isActive bool) ([]models.PaymentProviders, error)
	CreateStripeCheckoutSession(order *models.Order, payment *models.Payment, customerEmail string) (checkoutURL, sessionID string, err error)
	ConfirmStripeSession(sessionID, paymentIntentID string) (uint, error)
}

type paymentService struct {
	db           *gorm.DB
	stripe       *stripeintegration.Gateway
	stripeEnabled bool
}

func NewPaymentService(db *gorm.DB, cfg *config.Config) PaymentServiceInterface {
	svc := &paymentService{db: db, stripeEnabled: cfg.Stripe.Enabled}
	if cfg.Stripe.Enabled {
		svc.stripe = stripeintegration.NewGateway(cfg.Stripe.SecretKey, cfg.Email.FrontendURL)
	}
	return svc
}

func (s *paymentService) CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error) {
	tempTxID := fmt.Sprintf("pending_%d_%d", req.OrderID, time.Now().UnixNano())

	payment := &models.Payment{
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Method:        req.Method,
		TransactionID: tempTxID,
		Status:        "pending",
	}
	if err := tx.Create(payment).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return payment, nil
}

func (s *paymentService) ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error {
	var payment models.Payment
	if err := tx.First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if err := s.mockGateway(cardInfo); err != nil {
		payment.Status = "failed"
		payment.GatewayResponse = datatypes.JSON(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		if err := tx.Save(&payment).Error; err != nil {
			return fmt.Errorf("failed to save failed payment: %w", err)
		}
		return err
	}

	payment.Status = "succeeded"
	payment.TransactionID = fmt.Sprintf("txn_%d", time.Now().UnixNano())
	if err := tx.Save(&payment).Error; err != nil {
		return fmt.Errorf("failed to save successful payment: %w", err)
	}
	return nil
}

func (s *paymentService) CreateStripeCheckoutSession(order *models.Order, payment *models.Payment, customerEmail string) (string, string, error) {
	if !s.stripeEnabled || s.stripe == nil {
		return "", "", utils.ErrBadRequest("stripe payments are not enabled")
	}

	checkoutURL, sessionID, err := s.stripe.CreateCheckoutSession(order, payment.ID, customerEmail)
	if err != nil {
		return "", "", utils.ErrInternal(err)
	}

	payment.StripeSessionID = sessionID
	payment.Method = "stripe"
	if err := s.db.Model(payment).Updates(map[string]interface{}{
		"stripe_session_id": sessionID,
		"method":            "stripe",
	}).Error; err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return checkoutURL, sessionID, nil
}

// ConfirmStripeSession marks a payment as succeeded (idempotent). Returns the order ID.
func (s *paymentService) ConfirmStripeSession(sessionID, paymentIntentID string) (uint, error) {
	var payment models.Payment
	err := s.db.Where("stripe_session_id = ?", sessionID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, utils.ErrNotFound("payment not found for session")
		}
		return 0, utils.ErrInternal(err)
	}

	if payment.Status == "succeeded" {
		return payment.OrderID, nil
	}

	txnID := paymentIntentID
	if txnID == "" {
		txnID = sessionID
	}

	updates := map[string]interface{}{
		"status":         "succeeded",
		"transaction_id": txnID,
	}
	if err := s.db.Model(&payment).Updates(updates).Error; err != nil {
		return 0, utils.ErrInternal(err)
	}

	return payment.OrderID, nil
}

func (s *paymentService) mockGateway(cardInfo dto.CardInfo) error {
	time.Sleep(1 * time.Second)

	if cardInfo.CardNumber == "0000000000000000" {
		return errors.New("card declined")
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if cardInfo.ExpiryYear < year || (cardInfo.ExpiryYear == year && cardInfo.ExpiryMonth < month) {
		return errors.New("card expired")
	}

	if len(cardInfo.CVV) != 3 {
		return errors.New("invalid CVV")
	}

	return nil
}

func (s *paymentService) GetPaymentProvider(isActive bool) ([]models.PaymentProviders, error) {
	var methods []models.PaymentProviders
	err := s.db.Where("is_active = ?", isActive).
		Order("sort_order ASC").
		Find(&methods).Error
	return methods, err
}

func StripeEnabled(cfg *config.Config) bool {
	return cfg != nil && cfg.Stripe.Enabled
}
