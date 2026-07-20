package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domainpayment "github.com/alireza-akbarzadeh/luxe/internal/domain/payment"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/stripe/stripe-go/v82"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service orchestrates payment use cases.
type Service struct {
	repo          *postgres.PaymentRepository
	stripe        *stripeintegration.Gateway
	stripeEnabled bool
}

// NewService creates payment use cases.
func NewService(repo *postgres.PaymentRepository, stripeGateway *stripeintegration.Gateway, stripeEnabled bool) *Service {
	return &Service{repo: repo, stripe: stripeGateway, stripeEnabled: stripeEnabled}
}

func (s *Service) CreatePayment(tx *gorm.DB, req dto.PaymentRequest) (*models.Payment, error) {
	tempTxID := fmt.Sprintf("pending_%d_%d", req.OrderID, time.Now().UnixNano())

	payment := &models.Payment{
		OrderID:       req.OrderID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Method:        req.Method,
		TransactionID: tempTxID,
		Status:        constants.PaymentStatusPending,
	}
	if err := s.repo.CreatePaymentTx(tx, payment); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return payment, nil
}

func (s *Service) ProcessPayment(tx *gorm.DB, paymentID uint, cardInfo dto.CardInfo) error {
	payment, err := s.repo.FindPaymentTx(tx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	if err := s.mockGateway(cardInfo); err != nil {
		payment.Status = constants.PaymentStatusFailed
		payment.GatewayResponse = datatypes.JSON(fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		if err := s.repo.SavePaymentTx(tx, payment); err != nil {
			return fmt.Errorf("failed to save failed payment: %w", err)
		}
		return err
	}

	payment.Status = constants.PaymentStatusSucceeded
	payment.TransactionID = fmt.Sprintf("txn_%d", time.Now().UnixNano())
	if err := s.repo.SavePaymentTx(tx, payment); err != nil {
		return fmt.Errorf("failed to save successful payment: %w", err)
	}
	return nil
}

func (s *Service) CreateStripeCheckoutSession(order *models.Order, payment *models.Payment, customerEmail string) (string, string, error) {
	if !s.stripeEnabled || s.stripe == nil {
		return "", "", utils.ErrBadRequest("stripe payments are not enabled")
	}

	checkoutURL, sessionID, err := s.stripe.CreateCheckoutSession(order, payment.ID, customerEmail)
	if err != nil {
		return "", "", utils.ErrInternal(err)
	}

	if err := s.repo.UpdatePaymentFields(context.Background(), payment, map[string]interface{}{
		"stripe_session_id": sessionID,
		"method":            "stripe",
	}); err != nil {
		return "", "", utils.ErrInternal(err)
	}

	return checkoutURL, sessionID, nil
}

// GetStripeCheckoutSession loads a Stripe Checkout session by ID.
func (s *Service) GetStripeCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	if !s.stripeEnabled || s.stripe == nil {
		return nil, utils.ErrBadRequest("stripe payments are not configured")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, utils.ErrBadRequest("session_id is required")
	}
	return s.stripe.GetCheckoutSession(sessionID)
}

// FindByStripeSession loads a payment row linked to a Stripe Checkout session.
func (s *Service) FindByStripeSession(ctx context.Context, sessionID string) (*models.Payment, error) {
	payment, err := s.repo.FindByStripeSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("payment not found for session")
		}
		return nil, utils.ErrInternal(err)
	}
	return payment, nil
}

func (s *Service) ConfirmStripeSession(ctx context.Context, sessionID, paymentIntentID string) (uint, error) {
	payment, err := s.repo.FindByStripeSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, utils.ErrNotFound("payment not found for session")
		}
		return 0, utils.ErrInternal(err)
	}

	if payment.Status == constants.PaymentStatusSucceeded {
		return payment.OrderID, nil
	}

	txnID := paymentIntentID
	if txnID == "" {
		txnID = sessionID
	}

	updates := map[string]interface{}{
		"status":         constants.PaymentStatusSucceeded,
		"transaction_id": txnID,
	}
	if err := s.repo.UpdatePaymentByID(ctx, payment.ID, updates); err != nil {
		return 0, utils.ErrInternal(err)
	}

	return payment.OrderID, nil
}

func (s *Service) GetPaymentProvider(ctx context.Context, isActive bool) ([]models.PaymentProviders, error) {
	return s.repo.ListActiveProviders(ctx, isActive)
}

// ListAdmin returns paginated payments for the admin transactions list.
func (s *Service) ListAdmin(ctx context.Context, filters dto.AdminPaymentListFilters) ([]models.Payment, int64, error) {
	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filters.Page = page
	filters.Limit = limit
	offset := (page - 1) * limit

	total, err := s.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	payments, err := s.repo.ListAdmin(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return payments, total, nil
}

// GetByIDAdmin loads a payment with relations for the admin detail view.
func (s *Service) GetByIDAdmin(ctx context.Context, paymentID uint) (*models.Payment, error) {
	payment, err := s.repo.FindByIDAdmin(ctx, paymentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("payment not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return payment, nil
}

// GetSummaryAdmin returns admin KPI counters for payments.
func (s *Service) GetSummaryAdmin(ctx context.Context) (dto.PaymentsSummaryResponse, error) {
	total, err := s.repo.CountAll(ctx)
	if err != nil {
		return dto.PaymentsSummaryResponse{}, utils.ErrInternal(err)
	}
	statusCounts, err := s.repo.StatusCounts(ctx)
	if err != nil {
		return dto.PaymentsSummaryResponse{}, utils.ErrInternal(err)
	}
	volume, err := s.repo.SumAmountByStatuses(ctx, []string{constants.PaymentStatusSucceeded, constants.PaymentStatusCompleted})
	if err != nil {
		return dto.PaymentsSummaryResponse{}, utils.ErrInternal(err)
	}

	summary := dto.PaymentsSummaryResponse{TotalCount: total, TotalVolume: volume}
	for _, row := range statusCounts {
		switch row.Status {
		case constants.PaymentStatusSucceeded, constants.PaymentStatusCompleted:
			summary.CompletedCount += row.Count
		case constants.PaymentStatusPending:
			summary.PendingCount += row.Count
		case constants.PaymentStatusFailed:
			summary.FailedCount += row.Count
		case constants.PaymentStatusRefunded:
			summary.RefundedCount += row.Count
		}
	}
	return summary, nil
}

func (s *Service) mockGateway(cardInfo dto.CardInfo) error {
	time.Sleep(1 * time.Second)

	err := domainpayment.ValidateCard(domainpayment.Card{
		Number:      cardInfo.CardNumber,
		ExpiryMonth: cardInfo.ExpiryMonth,
		ExpiryYear:  cardInfo.ExpiryYear,
		CVV:         cardInfo.CVV,
	}, time.Now())
	if err != nil {
		return err
	}
	return nil
}
