package wallet

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domainwallet "github.com/alireza-akbarzadeh/luxe/internal/domain/wallet"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/stripe/stripe-go/v82"
	"gorm.io/gorm"
)

const walletDepositStripeMetadataType = "wallet_deposit"

// Notifier sends in-app notifications to users.
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// Service orchestrates wallet use cases.
type Service struct {
	repo          *postgres.WalletRepository
	stripe        *stripeintegration.Gateway
	stripeEnabled bool
	notifier      Notifier
}

// NewService creates wallet use cases.
func NewService(repo *postgres.WalletRepository, stripeGateway *stripeintegration.Gateway, stripeEnabled bool) *Service {
	return &Service{repo: repo, stripe: stripeGateway, stripeEnabled: stripeEnabled}
}

// SetNotifier injects the notification service (called from bootstrap after wiring).
func (w *Service) SetNotifier(notifier Notifier) {
	w.notifier = notifier
}

func (w *Service) GetOrCreateWallet(ctx context.Context, userID uint) (*models.Wallet, error) {
	wallet, err := w.repo.FindWalletByUserID(ctx, userID)
	if err == nil {
		return wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}
	newWallet := models.Wallet{
		UserID:   userID,
		Balance:  0,
		Currency: "USD",
	}
	if err := w.repo.CreateWallet(ctx, &newWallet); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &newWallet, nil
}

func (w *Service) GetBalance(ctx context.Context, userID uint) (float64, error) {
	wallet, err := w.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return 0, utils.ErrInternal(err)
	}
	return wallet.Balance, nil
}

func (w *Service) GetTransactions(ctx context.Context, userID uint, filters dto.WalletListFilters) ([]models.WalletTransaction, int64, error) {
	transactions, total, err := w.repo.ListTransactions(ctx, userID, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return transactions, total, nil
}

func (w *Service) Deposit(ctx context.Context, userID uint, amount float64, description string) error {
	if err := domainwallet.ValidatePositiveAmount(amount); err != nil {
		return utils.ErrBadRequest("amount must be positive")
	}
	return w.updateBalance(ctx, userID, amount, constants.WalletTxTypeDeposit, "", nil, description, constants.WalletTxStatusCompleted)
}

func (w *Service) CreatePendingDeposit(ctx context.Context, userID uint, amount float64, description string) (uint, error) {
	var txID uint
	err := w.repo.WithTx(func(tx *gorm.DB) error {
		record := models.WalletTransaction{
			UserID:       userID,
			Amount:       amount,
			Type:         constants.WalletTxTypeDeposit,
			Description:  description,
			BalanceAfter: 0,
			Status:       constants.WalletTxStatusPending,
		}
		if err := w.repo.CreateTransactionTx(tx, &record); err != nil {
			return err
		}
		txID = record.ID
		return nil
	})
	if err != nil {
		return 0, utils.ErrInternal(err)
	}
	return txID, nil
}

func (w *Service) InitiateDeposit(ctx context.Context, userID uint, amount float64, customerEmail string) (*dto.DepositResponse, error) {
	if err := domainwallet.ValidatePositiveAmount(amount); err != nil {
		return nil, utils.ErrBadRequest("amount must be positive")
	}
	txID, err := w.CreatePendingDeposit(ctx, userID, amount, "Online deposit via payment gateway")
	if err != nil {
		return nil, err
	}

	if !w.stripeEnabled || w.stripe == nil {
		if err := w.ConfirmDeposit(ctx, txID); err != nil {
			return nil, err
		}
		return &dto.DepositResponse{
			TransactionID: txID,
			Status:        constants.WalletTxStatusCompleted,
		}, nil
	}

	if customerEmail == "" {
		_ = w.FailDeposit(ctx, txID)
		return nil, utils.ErrBadRequest("customer email is required for stripe deposit")
	}

	checkoutURL, sessionID, err := w.stripe.CreateWalletDepositSession(userID, txID, amount, "USD", customerEmail)
	if err != nil {
		_ = w.FailDeposit(ctx, txID)
		return nil, utils.ErrInternal(err)
	}

	if err := w.repo.UpdateTransactionFields(ctx, txID, map[string]interface{}{
		"stripe_session_id": sessionID,
	}); err != nil {
		_ = w.FailDeposit(ctx, txID)
		return nil, utils.ErrInternal(err)
	}

	return &dto.DepositResponse{
		TransactionID:   txID,
		Status:          constants.WalletTxStatusPending,
		CheckoutURL:     checkoutURL,
		StripeSessionID: sessionID,
	}, nil
}

// ConfirmDepositBySessionID credits the wallet using the Stripe session ID from the success redirect.
func (w *Service) ConfirmDepositBySessionID(ctx context.Context, userID uint, sessionID string) (dto.ConfirmWalletDepositResponse, error) {
	if !w.stripeEnabled || w.stripe == nil {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrBadRequest("card payment is not configured")
	}

	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrBadRequest("session_id is required")
	}

	checkoutSession, err := w.stripe.GetCheckoutSession(sessionID)
	if err != nil {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrBadRequest("invalid or expired checkout session")
	}

	if checkoutSession.Metadata == nil || checkoutSession.Metadata["type"] != walletDepositStripeMetadataType {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrBadRequest("not a wallet deposit checkout session")
	}

	ownerIDStr := checkoutSession.Metadata["user_id"]
	ownerID64, parseErr := strconv.ParseUint(ownerIDStr, 10, 64)
	if parseErr != nil || ownerID64 == 0 || uint(ownerID64) != userID {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrForbidden("checkout session does not belong to this account")
	}

	if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrBadRequest("payment is not completed yet — refresh in a moment")
	}

	paymentIntentID := ""
	if checkoutSession.PaymentIntent != nil {
		paymentIntentID = checkoutSession.PaymentIntent.ID
	}

	if err := w.ConfirmDepositByStripeSession(ctx, checkoutSession.ID, paymentIntentID); err != nil {
		return dto.ConfirmWalletDepositResponse{}, err
	}

	txRecord, err := w.repo.FindTransactionByStripeSession(ctx, checkoutSession.ID)
	if err != nil {
		return dto.ConfirmWalletDepositResponse{}, utils.ErrInternal(err)
	}

	balance, err := w.GetBalance(ctx, userID)
	if err != nil {
		return dto.ConfirmWalletDepositResponse{}, err
	}

	amount := float64(checkoutSession.AmountTotal) / 100
	currency := strings.ToUpper(string(checkoutSession.Currency))
	if currency == "" {
		currency = "USD"
	}

	paidAt := time.Unix(checkoutSession.Created, 0).UTC()

	return dto.ConfirmWalletDepositResponse{
		Balance:  balance,
		Currency: currency,
		Receipt: dto.WalletDepositReceipt{
			Amount:          amount,
			Currency:        currency,
			BalanceAfter:    txRecord.BalanceAfter,
			PaidAt:          &paidAt,
			StripeSessionID: checkoutSession.ID,
			Status:          constants.WalletTxStatusCompleted,
			TransactionID:   txRecord.ID,
		},
	}, nil
}

func (w *Service) ConfirmDepositByStripeSession(ctx context.Context, sessionID, paymentIntentID string) error {
	txRecord, err := w.repo.FindTransactionByStripeSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("wallet deposit not found for session")
		}
		return utils.ErrInternal(err)
	}

	if txRecord.Status == constants.WalletTxStatusCompleted {
		return nil
	}

	if txRecord.Status != constants.WalletTxStatusPending {
		return utils.ErrBadRequest("transaction already processed")
	}

	if paymentIntentID != "" {
		if err := w.repo.UpdateTransactionFields(ctx, txRecord.ID, map[string]interface{}{
			"description": fmt.Sprintf("Stripe deposit (%s)", paymentIntentID),
		}); err != nil {
			utils.Log.WithError(err).Warn("failed to update wallet deposit description")
		}
	}

	return w.ConfirmDeposit(ctx, txRecord.ID)
}

func (w *Service) FailDepositByStripeSession(ctx context.Context, sessionID string) error {
	txRecord, err := w.repo.FindTransactionByStripeSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return utils.ErrInternal(err)
	}

	if txRecord.Status != constants.WalletTxStatusPending {
		return nil
	}

	return w.FailDeposit(ctx, txRecord.ID)
}

func (w *Service) ConfirmDeposit(ctx context.Context, transactionID uint) error {
	var userID uint
	var amount, balanceAfter float64

	err := w.repo.WithTx(func(tx *gorm.DB) error {
		txRecord, err := w.repo.FindTransactionTx(tx, transactionID)
		if err != nil {
			return err
		}
		if txRecord.Status != constants.WalletTxStatusPending {
			return utils.ErrBadRequest("transaction already processed")
		}

		wallet, err := w.repo.LockWalletForUpdate(tx, txRecord.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newWallet := models.Wallet{UserID: txRecord.UserID, Balance: 0, Currency: "USD"}
				if err := w.repo.CreateWalletTx(tx, &newWallet); err != nil {
					return utils.ErrInternal(err)
				}
				wallet = &newWallet
			} else {
				return utils.ErrInternal(err)
			}
		}
		newBalance := wallet.Balance + txRecord.Amount
		wallet.Balance = newBalance
		if err := w.repo.SaveWalletTx(tx, wallet); err != nil {
			return err
		}
		txRecord.Status = constants.WalletTxStatusCompleted
		txRecord.BalanceAfter = newBalance
		userID = txRecord.UserID
		amount = txRecord.Amount
		balanceAfter = newBalance
		return w.repo.SaveTransactionTx(tx, txRecord)
	})
	if err != nil {
		return err
	}

	w.notifyDepositCompleted(userID, transactionID, amount, balanceAfter)
	return nil
}

func (w *Service) notifyDepositCompleted(userID, transactionID uint, amount, balanceAfter float64) {
	if w.notifier == nil {
		return
	}

	go func() {
		_ = w.notifier.CreateNotification(
			userID,
			constants.NotificationTypeWalletDeposit,
			"Wallet topped up",
			fmt.Sprintf("$%.2f was added to your wallet. New balance: $%.2f.", amount, balanceAfter),
			map[string]interface{}{
				"transaction_id": transactionID,
				"amount":         amount,
				"balance_after":  balanceAfter,
				"currency":       "USD",
				"account_tab":    "payment",
			},
		)
	}()
}

func (w *Service) FailDeposit(ctx context.Context, transactionID uint) error {
	if err := w.repo.UpdateTransactionStatus(ctx, transactionID, constants.WalletTxStatusFailed); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

func (w *Service) Withdraw(ctx context.Context, userID uint, amount float64, referenceType string, referenceID *uint, description string) error {
	if err := domainwallet.ValidatePositiveAmount(amount); err != nil {
		return utils.ErrBadRequest("amount must be positive")
	}
	return w.updateBalance(ctx, userID, -amount, constants.WalletTxTypeAdjustment, referenceType, referenceID, description, constants.WalletTxStatusCompleted)
}

func (w *Service) DeductForOrder(ctx context.Context, userID uint, amount float64, orderID uint) error {
	return w.updateBalance(ctx, userID, -amount, constants.WalletTxTypePayment, constants.WalletRefTypeOrder, &orderID, fmt.Sprintf("Payment for order #%d", orderID), constants.WalletTxStatusCompleted)
}

func (w *Service) Refund(ctx context.Context, userID uint, amount float64, orderID uint) error {
	return w.updateBalance(ctx, userID, amount, constants.WalletTxTypeRefund, constants.WalletRefTypeOrder, &orderID, fmt.Sprintf("Refund for order #%d", orderID), constants.WalletTxStatusCompleted)
}

func (w *Service) AdminAdjust(ctx context.Context, userID uint, amount float64, description string) error {
	return w.updateBalance(ctx, userID, amount, "adjustment", constants.WalletRefTypeAdmin, nil, description, constants.WalletTxStatusCompleted)
}

func (w *Service) updateBalance(ctx context.Context, userID uint, delta float64, txType, refType string, refID *uint, description string, status string) error {
	if status == "" {
		status = constants.WalletTxStatusCompleted
	}
	return w.repo.WithTx(func(tx *gorm.DB) error {
		wallet, err := w.repo.LockWalletForUpdate(tx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newWallet := models.Wallet{UserID: userID, Balance: 0, Currency: "USD"}
				if err := w.repo.CreateWalletTx(tx, &newWallet); err != nil {
					return utils.ErrInternal(err)
				}
				wallet = &newWallet
			} else {
				return utils.ErrInternal(err)
			}
		}
		newBalance := wallet.Balance + delta
		if err := domainwallet.CanApplyDelta(wallet.Balance, delta); err != nil {
			return utils.ErrBadRequest("insufficient wallet balance")
		}
		wallet.Balance = newBalance
		if err := w.repo.SaveWalletTx(tx, wallet); err != nil {
			return err
		}
		record := models.WalletTransaction{
			UserID:        userID,
			Amount:        delta,
			Type:          txType,
			ReferenceType: refType,
			ReferenceID:   refID,
			Description:   description,
			BalanceAfter:  newBalance,
			Status:        status,
		}
		return w.repo.CreateTransactionTx(tx, &record)
	})
}

func (w *Service) GetTransaction(ctx context.Context, userID, txID uint) (*models.WalletTransaction, error) {
	txRecord, err := w.repo.FindTransactionByID(ctx, userID, txID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("transaction not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return txRecord, nil
}

func (w *Service) CancelPendingDeposit(ctx context.Context, userID, txID uint) error {
	return w.repo.WithTx(func(tx *gorm.DB) error {
		record, err := w.repo.FindTransactionForUserTx(tx, userID, txID)
		if err != nil {
			return err
		}
		if record.Status != constants.WalletTxStatusPending {
			return utils.ErrBadRequest("only pending deposits can be cancelled")
		}
		if record.Type != constants.WalletTxTypeDeposit {
			return utils.ErrBadRequest("only deposit transactions can be cancelled")
		}
		record.Status = constants.WalletTxStatusCancelled
		return w.repo.SaveTransactionTx(tx, record)
	})
}
