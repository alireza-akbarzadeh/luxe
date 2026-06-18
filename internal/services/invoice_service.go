package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

// InvoiceServiceInterface manages customer invoices for admin billing operations.
type InvoiceServiceInterface interface {
	ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error)
	GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error)
	UpdateStatus(ctx context.Context, invoiceID uint, status string) error
}

type invoiceService struct {
	db *gorm.DB
}

func NewInvoiceService(db *gorm.DB) InvoiceServiceInterface {
	return &invoiceService{db: db}
}

func (s *invoiceService) ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	q := s.db.WithContext(ctx).Model(&models.Invoice{})
	q = s.applyInvoiceListFilters(q, filters)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	listQ := s.applyInvoiceListFilters(s.db.WithContext(ctx).Model(&models.Invoice{}), filters)
	var invoices []models.Invoice
	if err := listQ.
		Preload("Order").
		Preload("User").
		Preload("Payment").
		Order("invoices.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return invoices, total, nil
}

func (s *invoiceService) applyInvoiceListFilters(query *gorm.DB, filters dto.AdminInvoiceListFilters) *gorm.DB {
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

func (s *invoiceService) GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := s.db.WithContext(ctx).
		Preload("Order.Items.Product.Category").
		Preload("User").
		Preload("Payment").
		First(&invoice, invoiceID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("invoice not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return &invoice, nil
}

var allowedInvoiceStatuses = map[string]bool{
	constants.InvoiceStatusDraft:    true,
	constants.InvoiceStatusIssued:   true,
	constants.InvoiceStatusPaid:     true,
	constants.InvoiceStatusVoid:     true,
	constants.InvoiceStatusRefunded: true,
}

func (s *invoiceService) UpdateStatus(ctx context.Context, invoiceID uint, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if !allowedInvoiceStatuses[status] {
		return utils.ErrBadRequest("invalid invoice status")
	}

	var invoice models.Invoice
	if err := s.db.WithContext(ctx).First(&invoice, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("invoice not found")
		}
		return utils.ErrInternal(err)
	}

	updates := map[string]interface{}{
		"status": status,
	}
	now := time.Now()
	switch status {
	case constants.InvoiceStatusIssued:
		if invoice.IssuedAt == nil {
			updates["issued_at"] = now
		}
	case constants.InvoiceStatusPaid:
		if invoice.PaidAt == nil {
			updates["paid_at"] = now
		}
		if invoice.IssuedAt == nil {
			updates["issued_at"] = now
		}
	}

	if err := s.db.WithContext(ctx).Model(&invoice).Updates(updates).Error; err != nil {
		return utils.ErrInternal(fmt.Errorf("update invoice status: %w", err))
	}
	return nil
}
