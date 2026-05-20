package services

import (
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
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
	ConfirmDeposit(transactionID uint) error
	FailDeposit(transactionID uint) error
	Withdraw(userID uint, amount float64, referenceType string, referenceID *uint, description string) error
	DeductForOrder(userID uint, amount float64, orderID uint) error
	Refund(userID uint, amount float64, orderID uint) error
	AdminAdjust(userID uint, amount float64, description string) error
	GetTransaction(userID, txID uint) (*models.WalletTransaction, error)
	CancelPendingDeposit(userID, txID uint) error
}

type walletService struct {
	db *gorm.DB
}

func NewWalletService(db *gorm.DB) WalletServiceInterface {
	return &walletService{
		db: db,
	}
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
		// Lock wallet row
		var wallet models.Wallet
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", txRecord.UserID).First(&wallet).Error; err != nil {
			return err
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
