package services

import (
	"context"

	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	stripeintegration "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/stripe"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
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
	FailDepositByStripeSession(sessionID string) error
	FailDeposit(transactionID uint) error
	Withdraw(userID uint, amount float64, referenceType string, referenceID *uint, description string) error
	DeductForOrder(userID uint, amount float64, orderID uint) error
	Refund(userID uint, amount float64, orderID uint) error
	AdminAdjust(userID uint, amount float64, description string) error
	GetTransaction(userID, txID uint) (*models.WalletTransaction, error)
	CancelPendingDeposit(userID, txID uint) error
}

type walletService struct {
	app *appwallet.Service
}

func NewWalletService(db *gorm.DB, cfg *config.Config) WalletServiceInterface {
	stripeEnabled := StripeEnabled(cfg)
	var gateway *stripeintegration.Gateway
	if cfg != nil && cfg.Stripe.Enabled {
		gateway = stripeintegration.NewGateway(cfg.Stripe.SecretKey, cfg.Email.FrontendURL)
	}
	return &walletService{
		app: appwallet.NewService(postgres.NewWalletRepository(db), gateway, stripeEnabled),
	}
}

func (w *walletService) GetOrCreateWallet(userID uint) (*models.Wallet, error) {
	return w.app.GetOrCreateWallet(context.Background(), userID)
}

func (w *walletService) GetBalance(userID uint) (float64, error) {
	return w.app.GetBalance(context.Background(), userID)
}

func (w *walletService) GetTransactions(userID uint, filters dto.WalletListFilters) ([]models.WalletTransaction, int64, error) {
	return w.app.GetTransactions(context.Background(), userID, filters)
}

func (w *walletService) Deposit(userID uint, amount float64, description string) error {
	return w.app.Deposit(context.Background(), userID, amount, description)
}

func (w *walletService) CreatePendingDeposit(userID uint, amount float64, description string) (uint, error) {
	return w.app.CreatePendingDeposit(context.Background(), userID, amount, description)
}

func (w *walletService) InitiateDeposit(userID uint, amount float64, customerEmail string) (*dto.DepositResponse, error) {
	return w.app.InitiateDeposit(context.Background(), userID, amount, customerEmail)
}

func (w *walletService) ConfirmDepositByStripeSession(sessionID, paymentIntentID string) error {
	return w.app.ConfirmDepositByStripeSession(context.Background(), sessionID, paymentIntentID)
}

func (w *walletService) FailDepositByStripeSession(sessionID string) error {
	return w.app.FailDepositByStripeSession(context.Background(), sessionID)
}

func (w *walletService) ConfirmDeposit(transactionID uint) error {
	return w.app.ConfirmDeposit(context.Background(), transactionID)
}

func (w *walletService) FailDeposit(transactionID uint) error {
	return w.app.FailDeposit(context.Background(), transactionID)
}

func (w *walletService) Withdraw(userID uint, amount float64, referenceType string, referenceID *uint, description string) error {
	return w.app.Withdraw(context.Background(), userID, amount, referenceType, referenceID, description)
}

func (w *walletService) DeductForOrder(userID uint, amount float64, orderID uint) error {
	return w.app.DeductForOrder(context.Background(), userID, amount, orderID)
}

func (w *walletService) Refund(userID uint, amount float64, orderID uint) error {
	return w.app.Refund(context.Background(), userID, amount, orderID)
}

func (w *walletService) AdminAdjust(userID uint, amount float64, description string) error {
	return w.app.AdminAdjust(context.Background(), userID, amount, description)
}

func (w *walletService) GetTransaction(userID, txID uint) (*models.WalletTransaction, error) {
	return w.app.GetTransaction(context.Background(), userID, txID)
}

func (w *walletService) CancelPendingDeposit(userID, txID uint) error {
	return w.app.CancelPendingDeposit(context.Background(), userID, txID)
}
