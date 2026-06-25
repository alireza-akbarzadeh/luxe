package services

import (
	"context"

	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
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
	app *apppayment.Service
}

func NewPaymentService(db *gorm.DB, cfg *config.Config) PaymentServiceInterface {
	stripeEnabled := cfg != nil && cfg.Stripe.Enabled
	var gateway *stripeintegration.Gateway
	if stripeEnabled {
		gateway = stripeintegration.NewGateway(cfg.Stripe.SecretKey, cfg.Email.FrontendURL)
	}
	return &paymentService{
		app: apppayment.NewService(postgres.NewPaymentRepository(db), gateway, stripeEnabled),
	}
}

func (s *paymentService) CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error) {
	return s.app.CreatePayment(tx, req)
}

func (s *paymentService) ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error {
	return s.app.ProcessPayment(tx, paymentID, cardInfo)
}

func (s *paymentService) CreateStripeCheckoutSession(order *models.Order, payment *models.Payment, customerEmail string) (string, string, error) {
	return s.app.CreateStripeCheckoutSession(order, payment, customerEmail)
}

func (s *paymentService) ConfirmStripeSession(sessionID, paymentIntentID string) (uint, error) {
	return s.app.ConfirmStripeSession(context.Background(), sessionID, paymentIntentID)
}

func (s *paymentService) GetPaymentProvider(isActive bool) ([]models.PaymentProviders, error) {
	return s.app.GetPaymentProvider(context.Background(), isActive)
}

func StripeEnabled(cfg *config.Config) bool {
	return cfg != nil && cfg.Stripe.Enabled
}
