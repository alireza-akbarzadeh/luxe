package services

import (
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

type WalletServiceInterface interface {
	GetOrCreateWallet(userID uint) (*models.Wallet, error)
	GetBalance(userID uint) (float64, error)
	GetTransactions(userID uint, filters dto.WalletListFilters) ([]models.WalletTransaction, int64, error)
	Deposit(userID uint, amount float64, description string) error
	CreatePendingDeposit(userID uint, amount float64, description string) (uint, error)
	InitiateDeposit(userID uint, amount float64, customerEmail string) (*dto.DepositResponse, error)
	ConfirmDeposit(transactionID uint) error
	ConfirmDepositByStripeSession(sessionID, paymentIntentID string) error
	FailDeposit(transactionID uint) error
	Withdraw(userID uint, amount float64, referenceType string, referenceID *uint, description string) error
	DeductForOrder(userID uint, amount float64, orderID uint) error
	Refund(userID uint, amount float64, orderID uint) error
	AdminAdjust(userID uint, amount float64, description string) error
	GetTransaction(userID, txID uint) (*models.WalletTransaction, error)
	CancelPendingDeposit(userID, txID uint) error
}

type walletService struct {
	db            *gorm.DB
	stripe        *stripeintegration.Gateway
	stripeEnabled bool
}

func NewWalletService(db *gorm.DB, cfg *config.Config) WalletServiceInterface {
	svc := &walletService{db: db, stripeEnabled: StripeEnabled(cfg)}
	if cfg != nil && cfg.Stripe.Enabled {
		svc.stripe = stripeintegration.NewGateway(cfg.Stripe.SecretKey, cfg.Email.FrontendURL)
	}
	return svc
}

// GetOrCreateWallet – ensures a wallet exists for the user.
func (w *walletService) GetOrCreateWallet(userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := w.db.Where("user_id = ?", userID).First(&wallet).Error
	if err == nil {
		return &wallet, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}
	wallet = models.Wallet{
		UserID:   userID,
		Balance:  0,
		Currency: "USD",
	}
	if err := w.db.Create(&wallet).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return &wallet, nil
}

// GetBalance get user balance
func (w *walletService) GetBalance(userID uint) (float64, error) {
	wallet, err := w.GetOrCreateWallet(userID)
	if err != nil {
		return 0, utils.ErrInternal(err)
	}
	return wallet.Balance, nil
}

// GetTransactions get transactions for the users
func (w *walletService) GetTransactions(userID uint, filters dto.WalletListFilters) ([]models.WalletTransaction, int64, error) {
	var transactions []models.WalletTransaction
	var total int64
	limit := filters.Limit
	offset := filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := w.db.Model(&models.WalletTransaction{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return transactions, total, nil
}

// Deposit – immediately credit the wallet (for direct admin deposit or internal use)
func (w *walletService) Deposit(userID uint, amount float64, description string) error {
	return w.updateBalance(userID, amount, "deposit", "", nil, description, "completed")
}

// CreatePendingDeposit – immediately credit the wallet (for direct admin deposit or internal use)
func (w *walletService) CreatePendingDeposit(userID uint, amount float64, description string) (uint, error) {
	var txID uint
	err := w.db.Transaction(func(tx *gorm.DB) error {
		record := models.WalletTransaction{
			UserID:       userID,
			Amount:       amount,
			Type:         "deposit",
			Description:  description,
			BalanceAfter: 0,
			Status:       "pending",
		}
		if err := tx.Create(&record).Error; err != nil {
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

// InitiateDeposit creates a pending deposit and returns a Stripe Checkout URL when enabled.
func (w *walletService) InitiateDeposit(userID uint, amount float64, customerEmail string) (*dto.DepositResponse, error) {
	txID, err := w.CreatePendingDeposit(userID, amount, "Online deposit via payment gateway")
	if err != nil {
		return nil, err
	}

	if !w.stripeEnabled || w.stripe == nil {
		if err := w.ConfirmDeposit(txID); err != nil {
			return nil, err
		}
		return &dto.DepositResponse{
			TransactionID: txID,
			Status:        "completed",
		}, nil
	}

	if customerEmail == "" {
		w.FailDeposit(txID)
		return nil, utils.ErrBadRequest("customer email is required for stripe deposit")
	}

	checkoutURL, sessionID, err := w.stripe.CreateWalletDepositSession(userID, txID, amount, "USD", customerEmail)
	if err != nil {
		w.FailDeposit(txID)
		return nil, utils.ErrInternal(err)
	}

	if err := w.db.Model(&models.WalletTransaction{}).Where("id = ?", txID).Updates(map[string]interface{}{
		"stripe_session_id": sessionID,
	}).Error; err != nil {
		w.FailDeposit(txID)
		return nil, utils.ErrInternal(err)
	}

	return &dto.DepositResponse{
		TransactionID:   txID,
		Status:          "pending",
		CheckoutURL:     checkoutURL,
		StripeSessionID: sessionID,
	}, nil
}

// ConfirmDepositByStripeSession completes a wallet deposit after Stripe payment (idempotent).
func (w *walletService) ConfirmDepositByStripeSession(sessionID, paymentIntentID string) error {
	var txRecord models.WalletTransaction
	err := w.db.Where("stripe_session_id = ?", sessionID).First(&txRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("wallet deposit not found for session")
		}
		return utils.ErrInternal(err)
	}

	if txRecord.Status == "completed" {
		return nil
	}

	if txRecord.Status != "pending" {
		return utils.ErrBadRequest("transaction already processed")
	}

	if paymentIntentID != "" {
		if err := w.db.Model(&txRecord).Update("description", fmt.Sprintf("Stripe deposit (%s)", paymentIntentID)).Error; err != nil {
			utils.Log.WithError(err).Warn("failed to update wallet deposit description")
		}
	}

	return w.ConfirmDeposit(txRecord.ID)
}

// ConfirmDeposit – completes a pending deposit, updates wallet balance
func (w *walletService) ConfirmDeposit(transactionID uint) error {
	return w.db.Transaction(func(tx *gorm.DB) error {
		var txRecord models.WalletTransaction
		if err := tx.First(&txRecord, transactionID).Error; err != nil {
			return err
		}
		if txRecord.Status != "pending" {
			return utils.ErrBadRequest("transaction already processed")
		}
		// Lock wallet row (create on first deposit if missing).
		var wallet models.Wallet
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", txRecord.UserID).First(&wallet).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = models.Wallet{UserID: txRecord.UserID, Balance: 0, Currency: "USD"}
				if err := tx.Create(&wallet).Error; err != nil {
					return utils.ErrInternal(err)
				}
			} else {
				return utils.ErrInternal(err)
			}
		}
		newBalance := wallet.Balance + txRecord.Amount
		wallet.Balance = newBalance
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}
		txRecord.Status = "completed"
		txRecord.BalanceAfter = newBalance
		return tx.Save(&txRecord).Error
	})
}

// FailDeposit – marks a pending deposit as failed
func (w *walletService) FailDeposit(transactionID uint) error {
	return w.db.Model(&models.WalletTransaction{}).Where("id = ?", transactionID).Update("status", "failed").Error
}

// Withdraw – generic withdrawal (e.g., admin deduction)
func (w *walletService) Withdraw(userID uint, amount float64, referenceType string, referenceID *uint, description string) error {
	return w.updateBalance(userID, -amount, "adjustment", referenceType, referenceID, description, "completed")
}

// DeductForOrder – payment from wallet during checkout
func (w *walletService) DeductForOrder(userID uint, amount float64, orderID uint) error {
	return w.updateBalance(userID, -amount, "payment", "order", &orderID, fmt.Sprintf("Payment for order #%d", orderID), "completed")
}

// Refund – refund a payment back to wallet
func (w *walletService) Refund(userID uint, amount float64, orderID uint) error {
	return w.updateBalance(userID, amount, "refund", "order", &orderID, fmt.Sprintf("Refund for order #%d", orderID), "completed")
}

// AdminAdjust – direct adjustment (positive or negative) with custom description
func (w *walletService) AdminAdjust(userID uint, amount float64, description string) error {
	txType := "adjustment"
	if amount > 0 {
		// Could be treated as deposit, but keep as adjustment for audit
	}
	return w.updateBalance(userID, amount, txType, "admin", nil, description, "completed")
}

// internal helper – core balance update with transaction
func (w *walletService) updateBalance(userID uint, delta float64, txType, refType string, refID *uint, description string, status string) error {
	if status == "" {
		status = "completed"
	}
	return w.db.Transaction(func(tx *gorm.DB) error {
		// Lock wallet row for update
		var wallet models.Wallet
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create wallet if not exists
				wallet = models.Wallet{UserID: userID, Balance: 0, Currency: "USD"}
				if err := tx.Create(&wallet).Error; err != nil {
					return utils.ErrInternal(err)
				}
			} else {
				return utils.ErrInternal(err)
			}
		}
		newBalance := wallet.Balance + delta
		if newBalance < 0 {
			return utils.ErrBadRequest("insufficient wallet balance")
		}
		wallet.Balance = newBalance
		if err := tx.Save(&wallet).Error; err != nil {
			return utils.ErrInternal(err)
		}
		// Create transaction record
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
		if err := tx.Create(&record).Error; err != nil {
			return utils.ErrInternal(err)
		}
		return nil
	})
}

// GetTransaction get transaction with given id
func (w *walletService) GetTransaction(userID, txID uint) (*models.WalletTransaction, error) {
	var tx models.WalletTransaction
	err := w.db.Where("id = ? AND user_id = ?", txID, userID).First(&tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("transaction not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &tx, nil
}

// CancelPendingDeposit canceling the pending request
func (w *walletService) CancelPendingDeposit(userID, txID uint) error {
	return w.db.Transaction(func(tx *gorm.DB) error {
		var record models.WalletTransaction
		if err := tx.Where("id = ? AND user_id = ?", txID, userID).First(&record).Error; err != nil {
			return err
		}
		if record.Status != "pending" {
			return utils.ErrBadRequest("only pending deposits can be cancelled")
		}
		if record.Type != "deposit" {
			return utils.ErrBadRequest("only deposit transactions can be cancelled")
		}
		record.Status = "cancelled"
		return tx.Save(&record).Error
	})
}
