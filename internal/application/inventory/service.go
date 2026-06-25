package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/asynq"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

type backInStockNotifier interface {
	NotifyBackInStock(productID uint, productName, productSlug string) error
}

// Notifier sends in-app notifications for inventory alerts.
type Notifier interface {
	CreateNotification(userID uint, notificationType, title, message string, data interface{}) error
}

// Service orchestrates inventory ledger and stock alert use cases.
type Service struct {
	engine        *workflow.Engine
	notifier      backInStockNotifier
	notifications Notifier
	jobQueue      asynq.JobQueue
	alertEmails   []string
	commands      *Commands
	queries       *Queries
}

// NewService wires inventory commands and queries.
func NewService(
	db *gorm.DB,
	engine *workflow.Engine,
	notifier backInStockNotifier,
	notifications Notifier,
	jobQueue asynq.JobQueue,
	alertEmails []string,
) *Service {
	repo := postgres.NewInventoryRepository(db)
	return &Service{
		engine:        engine,
		notifier:      notifier,
		notifications: notifications,
		jobQueue:      jobQueue,
		alertEmails:   alertEmails,
		commands:      NewCommands(repo),
		queries:       NewQueries(repo, postgres.RevenueOrderStatuses()),
	}
}

func (s *Service) GetOverview(ctx context.Context) (*dto.InventoryOverviewResponse, error) {
	overview, err := s.queries.GetOverview(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return overview, nil
}

func (s *Service) List(ctx context.Context, req *dto.ListInventoryRequest) ([]dto.InventoryItemResponse, int64, error) {
	items, total, err := s.queries.List(ctx, req)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return items, total, nil
}

func (s *Service) AdjustStock(ctx context.Context, actorUserID *uint, req *dto.AdjustInventoryRequest) (*dto.InventoryItemResponse, error) {
	result, err := s.commands.AdjustStock(ctx, actorUserID, req)
	if err != nil {
		return nil, err
	}

	s.handleStockSideEffects(ctx, result.Product, result.QuantityBefore, result.QuantityAfter)

	item, err := s.queries.GetItem(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return item, nil
}

func (s *Service) ListHistory(ctx context.Context, productID uint, req *dto.ListInventoryHistoryRequest) ([]dto.InventoryAdjustmentResponse, int64, error) {
	rows, total, err := s.queries.ListHistory(ctx, productID, req)
	if err != nil {
		return nil, 0, utils.ErrInternal(err)
	}
	return rows, total, nil
}

func (s *Service) ListRecentAdjustments(ctx context.Context, limit int) ([]dto.InventoryAdjustmentResponse, error) {
	rows, err := s.queries.ListRecentAdjustments(ctx, limit)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	return rows, nil
}

func (s *Service) ApplyDelta(ctx context.Context, tx *gorm.DB, params DeltaParams) error {
	_, err := s.commands.ApplyDeltaLocked(ctx, tx, params)
	return err
}

func (s *Service) RunStockSideEffects(ctx context.Context, product models.Product, before, after int) {
	s.handleStockSideEffects(ctx, product, before, after)
}

func (s *Service) SetAbsoluteStock(ctx context.Context, productID uint, newStock int, actorUserID *uint, note, adjustmentType string) error {
	result, err := s.commands.SetAbsoluteStock(ctx, productID, newStock, actorUserID, note, adjustmentType)
	if err != nil {
		return err
	}
	if result.QuantityBefore != result.QuantityAfter {
		s.handleStockSideEffects(ctx, result.Product, result.QuantityBefore, result.QuantityAfter)
	}
	return nil
}

func (s *Service) RecordInitialStock(ctx context.Context, productID uint, quantity int) error {
	return s.commands.RecordInitialStock(ctx, productID, quantity)
}

func (s *Service) DecrementForSale(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (DeltaResult, error) {
	return s.commands.DecrementForSale(ctx, tx, orderID, productID, quantity)
}

func (s *Service) RestoreForOrderCancel(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (DeltaResult, error) {
	return s.commands.RestoreForOrderCancel(ctx, tx, orderID, productID, quantity)
}

func (s *Service) handleStockSideEffects(ctx context.Context, product models.Product, before, after int) {
	if !product.TrackInventory || before == after {
		return
	}

	if before == 0 && after > 0 && s.notifier != nil {
		_ = s.notifier.NotifyBackInStock(product.ID, product.Name, product.Slug)
		if s.engine != nil {
			_ = appworkflow.ApplyEvent(ctx, s.engine, workflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityProduct,
				EntityID:    product.ID,
				Event:       "restock",
				ActorRole:   constants.RoleAdmin,
			})
		}
	}

	if after == 0 && before > 0 && s.engine != nil {
		_ = appworkflow.ApplyEvent(ctx, s.engine, workflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityProduct,
			EntityID:    product.ID,
			Event:       "mark_out_of_stock",
			ActorRole:   constants.RoleAdmin,
		})
	}
}

func (s *Service) BulkAdjustStock(
	ctx context.Context,
	actorUserID *uint,
	req *dto.BulkAdjustInventoryRequest,
) (*dto.BulkAdjustInventoryResponse, error) {
	resp := &dto.BulkAdjustInventoryResponse{
		Rows: make([]dto.BulkInventoryAdjustRowResult, 0, len(req.Rows)),
	}

	for _, row := range req.Rows {
		sku := strings.TrimSpace(row.SKU)
		if sku == "" {
			resp.Failed++
			resp.Rows = append(resp.Rows, dto.BulkInventoryAdjustRowResult{
				SKU:     row.SKU,
				Success: false,
				Message: "sku is required",
			})
			continue
		}
		if row.Delta == 0 {
			resp.Failed++
			resp.Rows = append(resp.Rows, dto.BulkInventoryAdjustRowResult{
				SKU:     sku,
				Success: false,
				Message: "delta cannot be zero",
			})
			continue
		}

		product, err := s.queries.FindProductBySKU(ctx, sku)
		if err != nil {
			resp.Failed++
			msg := "product not found"
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				msg = "lookup failed"
			}
			resp.Rows = append(resp.Rows, dto.BulkInventoryAdjustRowResult{
				SKU:     sku,
				Success: false,
				Message: msg,
			})
			continue
		}

		note := strings.TrimSpace(row.Note)
		if note == "" {
			note = req.Reason
		}

		item, err := s.AdjustStock(ctx, actorUserID, &dto.AdjustInventoryRequest{
			ProductID: product.ID,
			Delta:     row.Delta,
			Reason:    req.Reason,
			Note:      note,
		})
		if err != nil {
			resp.Failed++
			resp.Rows = append(resp.Rows, dto.BulkInventoryAdjustRowResult{
				SKU:     sku,
				Success: false,
				Message: err.Error(),
			})
			continue
		}

		resp.Applied++
		stock := item.Stock
		productID := item.ID
		resp.Rows = append(resp.Rows, dto.BulkInventoryAdjustRowResult{
			SKU:       sku,
			Success:   true,
			ProductID: &productID,
			Stock:     &stock,
		})
	}

	return resp, nil
}

func (s *Service) SendLowStockAlerts(ctx context.Context) error {
	products, err := s.queries.FindLowStockTracked(ctx)
	if err != nil {
		return utils.ErrInternal(err)
	}

	if len(products) == 0 {
		utils.Log.Info("inventory alert: no low-stock products")
		return nil
	}

	var outOfStock, lowStock []models.Product
	for _, p := range products {
		if p.Stock == 0 {
			outOfStock = append(outOfStock, p)
		} else {
			lowStock = append(lowStock, p)
		}
	}

	recipients, err := s.resolveAlertRecipients(ctx)
	if err != nil {
		return err
	}

	title := fmt.Sprintf("Inventory alert: %d SKU(s) need attention", len(products))
	body := buildLowStockEmailBody(lowStock, outOfStock)

	for _, admin := range recipients.admins {
		_ = s.notifications.CreateNotification(admin.ID, "low_stock_alert", title, body, map[string]interface{}{
			"low_stock_count":    len(lowStock),
			"out_of_stock_count": len(outOfStock),
		})
	}

	if s.jobQueue != nil {
		for _, email := range recipients.emails {
			_ = s.jobQueue.EnqueueSendEmail(ctx, email, title, body)
		}
	}

	for _, p := range products {
		utils.Log.WithFields(map[string]interface{}{
			"product_id": p.ID,
			"sku":        p.SKU,
			"stock":      p.Stock,
			"threshold":  p.LowStockThreshold,
		}).Warn("LOW STOCK ALERT")
	}

	return nil
}

type alertRecipients struct {
	admins []models.User
	emails []string
}

func (s *Service) resolveAlertRecipients(ctx context.Context) (alertRecipients, error) {
	admins, err := s.queries.FindAdmins(ctx)
	if err != nil {
		return alertRecipients{}, utils.ErrInternal(err)
	}

	emails := append([]string(nil), s.alertEmails...)
	if len(emails) == 0 {
		for _, admin := range admins {
			if admin.Email != "" {
				emails = append(emails, admin.Email)
			}
		}
	}
	return alertRecipients{admins: admins, emails: emails}, nil
}

func buildLowStockEmailBody(lowStock, outOfStock []models.Product) string {
	var b strings.Builder
	b.WriteString("<p>The following products need inventory attention:</p><ul>")
	for _, p := range outOfStock {
		fmt.Fprintf(&b, "<li><strong>%s</strong> (%s) — <span style=\"color:#dc2626\">OUT OF STOCK</span></li>", p.Name, p.SKU)
	}
	for _, p := range lowStock {
		fmt.Fprintf(&b, "<li><strong>%s</strong> (%s) — %d left (threshold %d)</li>", p.Name, p.SKU, p.Stock, p.LowStockThreshold)
	}
	b.WriteString("</ul><p>Review inventory in the admin dashboard.</p>")
	return b.String()
}

func (s *Service) RestockForReturn(ctx context.Context, returnID uint) error {
	changes, err := s.commands.RestockForReturn(ctx, returnID)
	if err != nil {
		return err
	}
	for _, change := range changes {
		s.handleStockSideEffects(ctx, change.Product, change.QuantityBefore, change.QuantityAfter)
	}
	return nil
}
