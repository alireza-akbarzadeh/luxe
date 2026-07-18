package invoice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/invoicepdf"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

// Queries orchestrates invoice read use cases.
type Queries struct {
	repo *postgres.InvoiceRepository
}

// NewQueries creates invoice query use cases.
func NewQueries(repo *postgres.InvoiceRepository) *Queries {
	return &Queries{repo: repo}
}

// ListAdmin returns paginated invoices for admin.
func (q *Queries) ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error) {
	limit, offset := filters.Limit, filters.Offset
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	total, err := q.repo.CountAdmin(ctx, filters)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	invoices, err := q.repo.ListAdmin(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return invoices, total, nil
}

// GetByIDAdmin loads an invoice with relations.
func (q *Queries) GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	invoice, err := q.repo.FindByIDAdmin(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("invoice not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return invoice, nil
}

// GeneratePDF renders an invoice PDF.
func (q *Queries) GeneratePDF(ctx context.Context, invoiceID uint) ([]byte, string, error) {
	invoice, err := q.GetByIDAdmin(ctx, invoiceID)
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
