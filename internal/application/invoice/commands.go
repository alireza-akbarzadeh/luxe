package invoice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

var allowedInvoiceStatuses = map[string]bool{
	constants.InvoiceStatusDraft:    true,
	constants.InvoiceStatusIssued:   true,
	constants.InvoiceStatusPaid:     true,
	constants.InvoiceStatusVoid:     true,
	constants.InvoiceStatusRefunded: true,
}

// Commands orchestrates invoice write use cases.
type Commands struct {
	repo        *postgres.InvoiceRepository
	jobQueue    asynq.JobQueue
	frontendURL string
}

// NewCommands creates invoice command use cases.
func NewCommands(repo *postgres.InvoiceRepository, jobQueue asynq.JobQueue, frontendURL string) *Commands {
	return &Commands{repo: repo, jobQueue: jobQueue, frontendURL: frontendURL}
}

// UpdateStatus updates invoice status with timestamp side effects.
func (c *Commands) UpdateStatus(ctx context.Context, invoiceID uint, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	if !allowedInvoiceStatuses[status] {
		return utils.ErrBadRequest("invalid invoice status")
	}

	invoice, err := c.repo.FindByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("invoice not found")
		}
		return utils.ErrInternal(err)
	}

	updates := map[string]interface{}{"status": status}
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

	if err := c.repo.UpdateFields(ctx, invoice, updates); err != nil {
		return utils.ErrInternal(fmt.Errorf("update invoice status: %w", err))
	}
	return nil
}

// CreateFromPaidOrder creates a paid invoice for an order (idempotent).
func (c *Commands) CreateFromPaidOrder(ctx context.Context, orderID uint) (*models.Invoice, error) {
	existing, err := c.repo.FindByOrderID(ctx, orderID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, utils.ErrInternal(err)
	}

	order, err := c.repo.FindOrderForInvoice(ctx, orderID)
	if err != nil {
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

	if err := c.repo.Create(ctx, &invoice); err != nil {
		return nil, utils.ErrInternal(err)
	}

	if c.jobQueue != nil {
		if err := c.enqueueInvoiceEmail(ctx, &invoice); err != nil {
			utils.Log.WithError(err).WithField("invoice_id", invoice.ID).Warn("failed to enqueue invoice email")
		}
	}

	return &invoice, nil
}

// SendToCustomer enqueues invoice email to the customer.
func (c *Commands) SendToCustomer(ctx context.Context, invoiceID uint) error {
	invoice, err := c.repo.FindByIDWithUser(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("invoice not found")
		}
		return utils.ErrInternal(err)
	}
	return c.enqueueInvoiceEmail(ctx, invoice)
}

func (c *Commands) enqueueInvoiceEmail(ctx context.Context, invoice *models.Invoice) error {
	to := strings.TrimSpace(invoice.BillingEmail)
	if to == "" && invoice.UserID != 0 {
		email, err := c.repo.FindUserEmail(ctx, invoice.UserID)
		if err == nil {
			to = email
		}
	}
	if to == "" {
		return utils.ErrBadRequest("invoice has no billing email")
	}
	if c.jobQueue == nil {
		return utils.ErrInternal(errors.New("email queue unavailable"))
	}

	subject := fmt.Sprintf("Your invoice %s", invoice.InvoiceNumber)
	body := buildInvoiceEmailHTML(invoice, c.frontendURL)
	return c.jobQueue.EnqueueSendEmail(ctx, to, subject, body)
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
