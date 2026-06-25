package postgres

import (
	"context"

	domainwallet "github.com/alireza-akbarzadeh/luxe/internal/domain/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// WalletRepository implements wallet persistence with GORM.
type WalletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a GORM-backed wallet repository.
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// GetByUserID implements domain/wallet.Repository.
func (r *WalletRepository) GetByUserID(ctx context.Context, userID uint) (*domainwallet.Wallet, error) {
	m, err := r.FindWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domainwallet.Wallet{
		UserID:       m.UserID,
		BalanceCents: int64(m.Balance * 100),
		Currency:     m.Currency,
	}, nil
}

// FindWalletByUserID loads a wallet row for a user.
func (r *WalletRepository) FindWalletByUserID(ctx context.Context, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

// CreateWallet inserts a new wallet.
func (r *WalletRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

// ListTransactions returns paginated wallet transactions for a user.
func (r *WalletRepository) ListTransactions(ctx context.Context, userID uint, filters dto.WalletListFilters) ([]models.WalletTransaction, int64, error) {
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
	query := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, 0, err
	}
	return transactions, total, nil
}

// FindTransactionByID loads a transaction owned by the user.
func (r *WalletRepository) FindTransactionByID(ctx context.Context, userID, txID uint) (*models.WalletTransaction, error) {
	var tx models.WalletTransaction
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", txID, userID).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

// FindTransactionByStripeSession loads a transaction by Stripe session ID.
func (r *WalletRepository) FindTransactionByStripeSession(ctx context.Context, sessionID string) (*models.WalletTransaction, error) {
	var txRecord models.WalletTransaction
	if err := r.db.WithContext(ctx).Where("stripe_session_id = ?", sessionID).First(&txRecord).Error; err != nil {
		return nil, err
	}
	return &txRecord, nil
}

// FindTransactionByIDOnly loads a transaction by primary key.
func (r *WalletRepository) FindTransactionByIDOnly(ctx context.Context, transactionID uint) (*models.WalletTransaction, error) {
	var txRecord models.WalletTransaction
	if err := r.db.WithContext(ctx).First(&txRecord, transactionID).Error; err != nil {
		return nil, err
	}
	return &txRecord, nil
}

// CreateTransaction inserts a wallet transaction row.
func (r *WalletRepository) CreateTransaction(ctx context.Context, record *models.WalletTransaction) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// UpdateTransactionFields updates selected columns on a transaction.
func (r *WalletRepository) UpdateTransactionFields(ctx context.Context, txID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.WalletTransaction{}).Where("id = ?", txID).Updates(updates).Error
}

// UpdateTransactionStatus sets status on a transaction.
func (r *WalletRepository) UpdateTransactionStatus(ctx context.Context, txID uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.WalletTransaction{}).Where("id = ?", txID).Update("status", status).Error
}

// SaveTransaction persists transaction changes.
func (r *WalletRepository) SaveTransaction(ctx context.Context, record *models.WalletTransaction) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// SaveWallet persists wallet changes.
func (r *WalletRepository) SaveWallet(ctx context.Context, wallet *models.Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

// WithTx runs fn inside a database transaction.
func (r *WalletRepository) WithTx(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// DB returns the underlying GORM handle for transactional helpers.
func (r *WalletRepository) DB() *gorm.DB {
	return r.db
}

// LockWalletForUpdate loads a wallet row with FOR UPDATE inside a transaction.
func (r *WalletRepository) LockWalletForUpdate(tx *gorm.DB, userID uint) (*models.Wallet, error) {
	var wallet models.Wallet
	err := tx.Set("gorm:query_option", "FOR UPDATE").Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// CreateWalletTx creates a wallet inside a transaction.
func (r *WalletRepository) CreateWalletTx(tx *gorm.DB, wallet *models.Wallet) error {
	return tx.Create(wallet).Error
}

// SaveWalletTx saves a wallet inside a transaction.
func (r *WalletRepository) SaveWalletTx(tx *gorm.DB, wallet *models.Wallet) error {
	return tx.Save(wallet).Error
}

// SaveTransactionTx saves a transaction inside a transaction.
func (r *WalletRepository) SaveTransactionTx(tx *gorm.DB, record *models.WalletTransaction) error {
	return tx.Save(record).Error
}

// CreateTransactionTx creates a transaction row inside a transaction.
func (r *WalletRepository) CreateTransactionTx(tx *gorm.DB, record *models.WalletTransaction) error {
	return tx.Create(record).Error
}

// FindTransactionTx loads a transaction inside a transaction.
func (r *WalletRepository) FindTransactionTx(tx *gorm.DB, transactionID uint) (*models.WalletTransaction, error) {
	var txRecord models.WalletTransaction
	if err := tx.First(&txRecord, transactionID).Error; err != nil {
		return nil, err
	}
	return &txRecord, nil
}

// FindTransactionForUserTx loads a user-owned transaction inside a transaction.
func (r *WalletRepository) FindTransactionForUserTx(tx *gorm.DB, userID, txID uint) (*models.WalletTransaction, error) {
	var record models.WalletTransaction
	if err := tx.Where("id = ? AND user_id = ?", txID, userID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}
