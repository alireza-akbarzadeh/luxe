package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PaymentRepository implements payment persistence with GORM.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a GORM-backed payment repository.
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// CreatePaymentTx inserts a payment inside an existing transaction.
func (r *PaymentRepository) CreatePaymentTx(tx *gorm.DB, payment *models.Payment) error {
	return tx.Create(payment).Error
}

// FindPaymentTx loads a payment inside a transaction.
func (r *PaymentRepository) FindPaymentTx(tx *gorm.DB, paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := tx.First(&payment, paymentID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// SavePaymentTx persists payment changes inside a transaction.
func (r *PaymentRepository) SavePaymentTx(tx *gorm.DB, payment *models.Payment) error {
	return tx.Save(payment).Error
}

// UpdatePaymentFields updates selected columns on a payment.
func (r *PaymentRepository) UpdatePaymentFields(ctx context.Context, payment *models.Payment, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(payment).Updates(updates).Error
}

// FindByStripeSession loads a payment by Stripe session ID.
func (r *PaymentRepository) FindByStripeSession(ctx context.Context, sessionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.db.WithContext(ctx).Where("stripe_session_id = ?", sessionID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdatePaymentByID updates payment columns by ID.
func (r *PaymentRepository) UpdatePaymentByID(ctx context.Context, paymentID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", paymentID).Updates(updates).Error
}

// ListActiveProviders returns active payment providers.
func (r *PaymentRepository) ListActiveProviders(ctx context.Context, isActive bool) ([]models.PaymentProviders, error) {
	var methods []models.PaymentProviders
	err := r.db.WithContext(ctx).Where("is_active = ?", isActive).
		Order("sort_order ASC").
		Find(&methods).Error
	return methods, err
}

// DB returns the underlying GORM handle.
func (r *PaymentRepository) DB() *gorm.DB {
	return r.db
}

// applyAdminListFilters scopes a payments query using admin list filters.
func (r *PaymentRepository) applyAdminListFilters(query *gorm.DB, filters dto.AdminPaymentListFilters) *gorm.DB {
	if filters.Status != "" {
		query = query.Where("payments.status = ?", filters.Status)
	}
	if filters.Method != "" {
		query = query.Where("payments.method = ?", filters.Method)
	}
	if filters.UserID != nil {
		query = query.Where("payments.user_id = ?", *filters.UserID)
	}
	if filters.OrderID != nil {
		query = query.Where("payments.order_id = ?", *filters.OrderID)
	}
	if filters.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", filters.DateFrom); err == nil {
			query = query.Where("payments.created_at >= ?", t)
		}
	}
	if filters.DateTo != "" {
		if t, err := time.Parse("2006-01-02", filters.DateTo); err == nil {
			query = query.Where("payments.created_at < ?", t.Add(24*time.Hour))
		}
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		query = query.
			Joins("LEFT JOIN orders ON orders.id = payments.order_id").
			Where(
				"payments.transaction_id ILIKE ? OR payments.stripe_session_id ILIKE ? OR orders.order_number ILIKE ?",
				term, term, term,
			)
	}
	return query
}

// CountAdmin counts payments matching admin filters.
func (r *PaymentRepository) CountAdmin(ctx context.Context, filters dto.AdminPaymentListFilters) (int64, error) {
	q := r.applyAdminListFilters(r.db.WithContext(ctx).Model(&models.Payment{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdmin returns paginated payments with relations for the admin transactions view.
func (r *PaymentRepository) ListAdmin(ctx context.Context, filters dto.AdminPaymentListFilters, limit, offset int) ([]models.Payment, error) {
	q := r.applyAdminListFilters(r.db.WithContext(ctx).Model(&models.Payment{}), filters)
	var payments []models.Payment
	err := q.
		Preload("Order").
		Preload("User").
		Order("payments.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error
	return payments, err
}

// FindByIDAdmin loads a payment with relations for the admin detail view.
func (r *PaymentRepository) FindByIDAdmin(ctx context.Context, paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("Order").
		Preload("User").
		First(&payment, paymentID).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// CountAll counts all payments regardless of filters.
func (r *PaymentRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.Payment{}).Count(&total).Error
	return total, err
}

// StatusCounts returns payment counts grouped by status.
func (r *PaymentRepository) StatusCounts(ctx context.Context) ([]dto.AdminDashboardStatusCount, error) {
	var rows []dto.AdminDashboardStatusCount
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error
	return rows, err
}

// SumAmountByStatuses sums payment amounts for the given statuses.
func (r *PaymentRepository) SumAmountByStatuses(ctx context.Context, statuses []string) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("status IN ?", statuses).
		Scan(&total).Error
	return total, err
}
