package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/integrations/invoicepdf"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

// InvoiceServiceInterface manages customer invoices for admin billing operations.
type InvoiceServiceInterface interface {
	ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error)
	GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error)
	UpdateStatus(ctx context.Context, invoiceID uint, status string) error
	CreateFromPaidOrder(ctx context.Context, orderID uint) (*models.Invoice, error)
	GeneratePDF(ctx context.Context, invoiceID uint) ([]byte, string, error)
	SendToCustomer(ctx context.Context, invoiceID uint) error
}

type invoiceService struct {
	db          *gorm.DB
	jobQueue    tasks.JobQueue
	frontendURL string
}

func NewInvoiceService(db *gorm.DB, jobQueue tasks.JobQueue, cfg *config.Config) InvoiceServiceInterface {
	frontendURL := ""
	if cfg != nil {
		frontendURL = strings.TrimRight(cfg.Email.FrontendURL, "/")
	}
	return &invoiceService{
		db:          db,
		jobQueue:    jobQueue,
		frontendURL: frontendURL,
	}
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

// CreateFromPaidOrder creates a paid invoice for an order (idempotent). Sends email when newly created.
func (s *invoiceService) CreateFromPaidOrder(ctx context.Context, orderID uint) (*models.Invoice, error) {
	var existing models.Invoice
	if err := s.db.WithContext(ctx).Where("order_id = ?", orderID).First(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	var order models.Order
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Payment").
		Preload("Shipment").
		Preload("Items").
		First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("order not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if order.Status != constants.OrderStatusPaid {
		return nil, nil
	}

	shippingAmount := 0.0
	if order.Shipment != nil {
		shippingAmount = order.Shipment.ShippingPrice
	}
	subtotal := order.TotalAmount - shippingAmount
	if subtotal < 0 {
		subtotal = order.TotalAmount
	}

	now := time.Now()
	issuedAt := &now
	paidAt := &now
	status := constants.InvoiceStatusPaid

	var paymentID *uint
	if order.Payment != nil {
		paymentID = &order.Payment.ID
		if order.Payment.Status != constants.PaymentStatusSucceeded {
			status = constants.InvoiceStatusIssued
			paidAt = nil
		}
	}

	billingName := strings.TrimSpace(order.User.FirstName + " " + order.User.LastName)
	invoice := models.Invoice{
		InvoiceNumber:  fmt.Sprintf("INV-%s", order.OrderNumber),
		OrderID:        order.ID,
		UserID:         order.UserID,
		PaymentID:      paymentID,
		Subtotal:       subtotal,
		TaxAmount:      0,
		ShippingAmount: shippingAmount,
		TotalAmount:    order.TotalAmount,
		Currency:       order.Currency,
		Status:         status,
		IssuedAt:       issuedAt,
		PaidAt:         paidAt,
		BillingName:    billingName,
		BillingEmail:   order.User.Email,
	}

	if err := s.db.WithContext(ctx).Create(&invoice).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if s.jobQueue != nil {
		if err := s.enqueueInvoiceEmail(ctx, &invoice); err != nil {
			utils.Log.WithError(err).WithField("invoice_id", invoice.ID).Warn("failed to enqueue invoice email")
		}
	}

	return &invoice, nil
}

func (s *invoiceService) GeneratePDF(ctx context.Context, invoiceID uint) ([]byte, string, error) {
	invoice, err := s.GetByIDAdmin(ctx, invoiceID)
	if err != nil {
		return nil, "", err
	}
	detail := dto.ToInvoiceDetail(invoice)
	data, err := invoicepdf.Generate(detail)
	if err != nil {
		return nil, "", utils.ErrInternal(err)
	}
	filename := fmt.Sprintf("%s.pdf", strings.ReplaceAll(invoice.InvoiceNumber, "/", "-"))
	return data, filename, nil
}

func (s *invoiceService) SendToCustomer(ctx context.Context, invoiceID uint) error {
	var invoice models.Invoice
	if err := s.db.WithContext(ctx).Preload("User").First(&invoice, invoiceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("invoice not found")
		}
		return utils.ErrInternal(err)
	}
	return s.enqueueInvoiceEmail(ctx, &invoice)
}

func (s *invoiceService) enqueueInvoiceEmail(ctx context.Context, invoice *models.Invoice) error {
	to := strings.TrimSpace(invoice.BillingEmail)
	if to == "" && invoice.UserID != 0 {
		var user models.User
		if err := s.db.WithContext(ctx).Select("email").First(&user, invoice.UserID).Error; err == nil {
			to = user.Email
		}
	}
	if to == "" {
		return utils.ErrBadRequest("invoice has no billing email")
	}
	if s.jobQueue == nil {
		return utils.ErrInternal(errors.New("email queue unavailable"))
	}

	subject := fmt.Sprintf("Your invoice %s", invoice.InvoiceNumber)
	body := buildInvoiceEmailHTML(invoice, s.frontendURL)
	return s.jobQueue.EnqueueSendEmail(ctx, to, subject, body)
}

func buildInvoiceEmailHTML(invoice *models.Invoice, frontendURL string) string {
	orderLink := ""
	if frontendURL != "" && invoice.OrderID > 0 {
		orderLink = fmt.Sprintf(`<p><a href="%s/account/orders/%d">View your order</a></p>`, frontendURL, invoice.OrderID)
	}

	paidLine := "Pending"
	if invoice.PaidAt != nil {
		paidLine = invoice.PaidAt.Format(time.RFC1123)
	}

	return fmt.Sprintf(`
		<h2>Invoice %s</h2>
		<p>Hello %s,</p>
		<p>Thank you for your purchase. Here is your invoice summary:</p>
		<ul>
			<li><strong>Invoice:</strong> %s</li>
			<li><strong>Status:</strong> %s</li>
			<li><strong>Total:</strong> %.2f %s</li>
			<li><strong>Paid:</strong> %s</li>
		</ul>
		%s
		<p>If you have questions, reply to this email or contact support.</p>
	`,
		invoice.InvoiceNumber,
		invoice.BillingName,
		invoice.InvoiceNumber,
		strings.ToUpper(invoice.Status),
		invoice.TotalAmount,
		invoice.Currency,
		paidLine,
		orderLink,
	)
}
