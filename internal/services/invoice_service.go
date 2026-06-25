package services

import (
	"context"
	"strings"

	appinvoice "github.com/alireza-akbarzadeh/luxe/internal/application/invoice"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

type InvoiceServiceInterface interface {
	ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error)
	GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error)
	UpdateStatus(ctx context.Context, invoiceID uint, status string) error
	CreateFromPaidOrder(ctx context.Context, orderID uint) (*models.Invoice, error)
	GeneratePDF(ctx context.Context, invoiceID uint) ([]byte, string, error)
	SendToCustomer(ctx context.Context, invoiceID uint) error
}

type invoiceService struct {
	queries  *appinvoice.Queries
	commands *appinvoice.Commands
}

func NewInvoiceService(db *gorm.DB, jobQueue asynq.JobQueue, cfg *config.Config) InvoiceServiceInterface {
	repo := postgres.NewInvoiceRepository(db)
	frontendURL := ""
	if cfg != nil {
		frontendURL = strings.TrimRight(cfg.Email.FrontendURL, "/")
	}
	return &invoiceService{
		queries:  appinvoice.NewQueries(repo),
		commands: appinvoice.NewCommands(repo, jobQueue, frontendURL),
	}
}

func (s *invoiceService) ListAdmin(ctx context.Context, filters dto.AdminInvoiceListFilters) ([]models.Invoice, int64, error) {
	return s.queries.ListAdmin(ctx, filters)
}

func (s *invoiceService) GetByIDAdmin(ctx context.Context, invoiceID uint) (*models.Invoice, error) {
	return s.queries.GetByIDAdmin(ctx, invoiceID)
}

func (s *invoiceService) UpdateStatus(ctx context.Context, invoiceID uint, status string) error {
	return s.commands.UpdateStatus(ctx, invoiceID, status)
}

func (s *invoiceService) CreateFromPaidOrder(ctx context.Context, orderID uint) (*models.Invoice, error) {
	return s.commands.CreateFromPaidOrder(ctx, orderID)
}

func (s *invoiceService) GeneratePDF(ctx context.Context, invoiceID uint) ([]byte, string, error) {
	return s.queries.GeneratePDF(ctx, invoiceID)
}

func (s *invoiceService) SendToCustomer(ctx context.Context, invoiceID uint) error {
	return s.commands.SendToCustomer(ctx, invoiceID)
}
