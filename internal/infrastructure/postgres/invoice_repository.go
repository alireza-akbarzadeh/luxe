package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// InvoiceRepository persists invoices with GORM.
type InvoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a GORM-backed invoice repository.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) applyListFilters(query *gorm.DB, filters dto.AdminInvoiceListFilters) *gorm.DB {
	if filters.Status != "" {
		query = query.Where("invoices.status = ?", filters.Status)
	}
	if filters.UserID != nil {
		query = query.Where("invoices.user_id = ?", *filters.UserID)
	}
	if filters.OrderID != nil {
		query = query.Where("invoices.order_id = ?", *filters.OrderID)
	}
	if filters.FromDate != "" {
		if t, err := time.Parse(time.RFC3339, filters.FromDate); err == nil {
			query = query.Where("invoices.created_at >= ?", t)
		}
	}
	if filters.ToDate != "" {
		if t, err := time.Parse(time.RFC3339, filters.ToDate); err == nil {
			query = query.Where("invoices.created_at <= ?", t)
		}
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		query = query.
			Joins("LEFT JOIN orders ON orders.id = invoices.order_id").
			Joins("LEFT JOIN users ON users.id = invoices.user_id").
			Where(
				"invoices.invoice_number ILIKE ? OR orders.order_number ILIKE ? OR users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?",
				term, term, term, term, term,
			)
	}
	return query
}

// CountAdmin counts invoices matching filters.
func (r *InvoiceRepository) CountAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) (int64, error) {
	q := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Invoice{}), filters)
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// ListAdmin returns paginated invoices for admin.
func (r *InvoiceRepository) ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters, limit, offset int) ([]models.Invoice, error) {
	listQ := r.applyListFilters(r.db.WithContext(ctx).Model(&models.Invoice{}), filters)
	var invoices []models.Invoice
	err := listQ.
		Preload("Order").
		Preload("User").
		Preload("Payment").
		Order("invoices.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error
	return invoices, err
}

// FindByIDAdmin loads an invoice with relations.
func (r *InvoiceRepository) FindByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).
		Preload("Order.Items.Product.Category").
		Preload("User").
		Preload("Payment").
		First(&invoice, invoiceID).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindByID loads a basic invoice row.
func (r *InvoiceRepository) FindByID(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).First(&invoice, invoiceID).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindByIDWithUser loads an invoice with user preload.
func (r *InvoiceRepository) FindByIDWithUser(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).Preload("User").First(&invoice, invoiceID).Error
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

// UpdateFields updates invoice columns.
func (r *InvoiceRepository) UpdateFields(ctx context.Context, invoice *models.Invoice, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(invoice).Updates(updates).Error
}

// FindByOrderID loads an invoice for an order if present.
func (r *InvoiceRepository) FindByOrderID(ctx context.Context, orderID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindOrderForInvoice loads an order with relations needed to create an invoice.
func (r *InvoiceRepository) FindOrderForInvoice(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Payment").
		Preload("Shipment").
		Preload("Items").
		First(&order, orderID).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Create inserts an invoice.
func (r *InvoiceRepository) Create(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// FindUserEmail loads a user's email by id.
func (r *InvoiceRepository) FindUserEmail(ctx context.Context, userID uint) (string, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Select("email").First(&user, userID).Error; err != nil {
		return "", err
	}
	return strings.TrimSpace(user.Email), nil
}
