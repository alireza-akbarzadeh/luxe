package postgres

import (
	"context"
	"time"

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

// applyAdminListFilters scopes a wallet transactions query using admin list filters.
func (r *WalletRepository) applyAdminListFilters(query *gorm.DB, filters dto.AdminWalletTxListFilters) *gorm.DB {
	if filters.Type != "" {
		query = query.Where("wallet_transactions.type = ?", filters.Type)
	}
	if filters.Status != "" {
		query = query.Where("wallet_transactions.status = ?", filters.Status)
	}
	if filters.UserID != nil {
		query = query.Where("wallet_transactions.user_id = ?", *filters.UserID)
	}
	if filters.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", filters.DateFrom); err == nil {
			query = query.Where("wallet_transactions.created_at >= ?", t)
		}
	}
	if filters.DateTo != "" {
		if t, err := time.Parse("2006-01-02", filters.DateTo); err == nil {
			query = query.Where("wallet_transactions.created_at < ?", t.Add(24*time.Hour))
		}
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		query = query.
			Joins("LEFT JOIN users ON users.id = wallet_transactions.user_id").
			Where(
				"wallet_transactions.description ILIKE ? OR wallet_transactions.stripe_session_id ILIKE ? OR users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?",
				term, term, term, term, term,
			)
	}
	return query
}

// CountAdminTransactions counts wallet transactions matching admin filters.
func (r *WalletRepository) CountAdminTransactions(ctx context.Context, filters dto.AdminWalletTxListFilters) (int64, error) {
	q := r.applyAdminListFilters(r.db.WithContext(ctx).Model(&models.WalletTransaction{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdminTransactions returns paginated wallet transactions with relations for the admin ledger view.
func (r *WalletRepository) ListAdminTransactions(ctx context.Context, filters dto.AdminWalletTxListFilters, limit, offset int) ([]models.WalletTransaction, error) {
	q := r.applyAdminListFilters(r.db.WithContext(ctx).Model(&models.WalletTransaction{}), filters)
	var transactions []models.WalletTransaction
	err := q.
		Preload("User").
		Order("wallet_transactions.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// FindTransactionByIDAdmin loads a wallet transaction with relations for the admin detail view.
func (r *WalletRepository) FindTransactionByIDAdmin(ctx context.Context, txID uint) (*models.WalletTransaction, error) {
	var record models.WalletTransaction
	err := r.db.WithContext(ctx).Preload("User").First(&record, txID).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// CountAllTransactions counts all wallet transactions regardless of filters.
func (r *WalletRepository) CountAllTransactions(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).Count(&total).Error
	return total, err
}

// TransactionStatusCounts returns wallet transaction counts grouped by status.
func (r *WalletRepository) TransactionStatusCounts(ctx context.Context) ([]dto.AdminDashboardStatusCount, error) {
	var rows []dto.AdminDashboardStatusCount
	err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error
	return rows, err
}

// TransactionTypeCounts returns wallet transaction counts grouped by type.
func (r *WalletRepository) TransactionTypeCounts(ctx context.Context) ([]dto.WalletTxTypeCount, error) {
	var rows []dto.WalletTxTypeCount
	err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&rows).Error
	return rows, err
}

// SumAbsAmountByStatus sums the absolute wallet transaction amounts for a status.
func (r *WalletRepository) SumAbsAmountByStatus(ctx context.Context, status string) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.WalletTransaction{}).
		Select("COALESCE(SUM(ABS(amount)), 0)").
		Where("status = ?", status).
		Scan(&total).Error
	return total, err
}
