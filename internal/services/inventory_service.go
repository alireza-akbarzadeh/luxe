package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/tasks"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type backInStockNotifier interface {
	NotifyBackInStock(productID uint, productName, productSlug string) error
}

// DeltaParams describes a stock change inside an existing transaction.
type DeltaParams struct {
	ProductID             uint
	Delta                 int
	AdjustmentType        string
	ReferenceType         string
	ReferenceID           *uint
	ActorUserID           *uint
	Note                  string
	SkipAvailabilityCheck bool
}

type deltaResult struct {
	Product       models.Product
	QuantityBefore int
	QuantityAfter  int
}

type InventoryServiceInterface interface {
	GetOverview(ctx context.Context) (*dto.InventoryOverviewResponse, error)
	List(ctx context.Context, req *dto.ListInventoryRequest) ([]dto.InventoryItemResponse, int64, error)
	AdjustStock(ctx context.Context, actorUserID *uint, req *dto.AdjustInventoryRequest) (*dto.InventoryItemResponse, error)
	ListHistory(ctx context.Context, productID uint, req *dto.ListInventoryHistoryRequest) ([]dto.InventoryAdjustmentResponse, int64, error)
	ListRecentAdjustments(ctx context.Context, limit int) ([]dto.InventoryAdjustmentResponse, error)
	ApplyDelta(ctx context.Context, tx *gorm.DB, params DeltaParams) error
	SetAbsoluteStock(ctx context.Context, productID uint, newStock int, actorUserID *uint, note, adjustmentType string) error
	RecordInitialStock(ctx context.Context, productID uint, quantity int) error
	RestoreForOrderCancel(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (deltaResult, error)
	DecrementForSale(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (deltaResult, error)
	RunStockSideEffects(ctx context.Context, product models.Product, before, after int)
	BulkAdjustStock(ctx context.Context, actorUserID *uint, req *dto.BulkAdjustInventoryRequest) (*dto.BulkAdjustInventoryResponse, error)
	SendLowStockAlerts(ctx context.Context) error
	RestockForReturn(ctx context.Context, returnID uint) error
}

type inventoryService struct {
	db            *gorm.DB
	engine        *workflow.Engine
	notifier      backInStockNotifier
	notifications NotificationServiceInterface
	jobQueue      tasks.JobQueue
	alertEmails   []string
}

func NewInventoryService(
	db *gorm.DB,
	engine *workflow.Engine,
	notifier backInStockNotifier,
	notifications NotificationServiceInterface,
	jobQueue tasks.JobQueue,
	alertEmails []string,
) InventoryServiceInterface {
	return &inventoryService{
		db:            db,
		engine:        engine,
		notifier:      notifier,
		notifications: notifications,
		jobQueue:      jobQueue,
		alertEmails:   alertEmails,
	}
}

func (s *inventoryService) GetOverview(ctx context.Context) (*dto.InventoryOverviewResponse, error) {
	db := s.db.WithContext(ctx)
	overview := &dto.InventoryOverviewResponse{}

	if err := db.Model(&models.Product{}).
		Where("track_inventory = ? AND stock <= low_stock_threshold AND stock > 0 AND status = ?",
			true, constants.ProductStatusActive).
		Count(&overview.LowStockCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if err := db.Model(&models.Product{}).
		Where("track_inventory = ? AND stock = 0 AND status = ?", true, constants.ProductStatusActive).
		Count(&overview.OutOfStockCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if err := db.Model(&models.Product{}).
		Where("track_inventory = ?", false).
		Count(&overview.NotTrackedCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	if err := db.Model(&models.Product{}).
		Where("track_inventory = ?", true).
		Count(&overview.TrackedSKUCount).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	type sumRow struct {
		Total int64
	}
	var sum sumRow
	if err := db.Model(&models.Product{}).
		Select("COALESCE(SUM(stock), 0) AS total").
		Where("track_inventory = ?", true).
		Scan(&sum).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	overview.TotalUnitsOnHand = sum.Total

	if err := db.Model(&models.StockNotification{}).
		Where("status = ?", constants.StockNotificationStatusActive).
		Count(&overview.WaitlistTotal).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}

	return overview, nil
}

type inventoryListRow struct {
	models.Product
	WaitlistCount int64 `gorm:"column:waitlist_count"`
	UnitsSold30d  int64 `gorm:"column:units_sold_30d"`
}

func (s *inventoryService) List(ctx context.Context, req *dto.ListInventoryRequest) ([]dto.InventoryItemResponse, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.WithContext(ctx).Model(&models.Product{}).
		Select(`
			products.*,
			COALESCE(waitlist.waitlist_count, 0) AS waitlist_count,
			COALESCE(sales.units_sold_30d, 0) AS units_sold_30d`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) AS waitlist_count
			FROM stock_notifications
			WHERE status = ?
			GROUP BY product_id
		) waitlist ON waitlist.product_id = products.id`, constants.StockNotificationStatusActive).
		Joins(`LEFT JOIN (
			SELECT oi.product_id, COALESCE(SUM(oi.quantity), 0) AS units_sold_30d
			FROM order_items oi
			INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL
			WHERE o.created_at >= ? AND o.status IN ?
			GROUP BY oi.product_id
		) sales ON sales.product_id = products.id`,
			time.Now().AddDate(0, 0, -30),
			revenueOrderStatuses,
		)

	if req.Search != "" {
		search := "%" + strings.TrimSpace(req.Search) + "%"
		query = query.Where("products.name ILIKE ? OR products.sku ILIKE ?", search, search)
	}

	switch req.StockStatus {
	case constants.InventoryStockLow:
		query = query.Where("products.track_inventory = ? AND products.stock > 0 AND products.stock <= products.low_stock_threshold", true)
	case constants.InventoryStockOut:
		query = query.Where("products.track_inventory = ? AND products.stock = 0", true)
	case constants.InventoryStockHealthy:
		query = query.Where("products.track_inventory = ? AND products.stock > products.low_stock_threshold", true)
	case constants.InventoryStockNotTracked:
		query = query.Where("products.track_inventory = ?", false)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	switch req.Sort {
	case "stock_asc":
		query = query.Order("products.stock ASC, products.name ASC")
	case "stock_desc":
		query = query.Order("products.stock DESC, products.name ASC")
	case "name_asc":
		query = query.Order("products.name ASC")
	case "waitlist_desc":
		query = query.Order("waitlist_count DESC, products.stock ASC")
	case "velocity_desc":
		query = query.Order("units_sold_30d DESC, products.stock ASC")
	default:
		query = query.Order("products.stock ASC, products.name ASC")
	}

	offset := (page - 1) * limit
	var rows []inventoryListRow
	if err := query.Offset(offset).Limit(limit).
		Preload("WorkflowState").
		Find(&rows).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	items := make([]dto.InventoryItemResponse, 0, len(rows))
	for i := range rows {
		items = append(items, toInventoryItemResponse(&rows[i].Product, rows[i].WaitlistCount, rows[i].UnitsSold30d))
	}
	return items, total, nil
}

func (s *inventoryService) AdjustStock(ctx context.Context, actorUserID *uint, req *dto.AdjustInventoryRequest) (*dto.InventoryItemResponse, error) {
	if req.Delta == 0 {
		return nil, utils.ErrBadRequest("adjustment delta cannot be zero")
	}

	adjType := mapReasonToAdjustmentType(req.Reason)
	note := strings.TrimSpace(req.Note)
	if note == "" {
		note = req.Reason
	}

	var result deltaResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = s.applyDeltaLocked(ctx, tx, DeltaParams{
			ProductID:      req.ProductID,
			Delta:          req.Delta,
			AdjustmentType: adjType,
			ReferenceType:  constants.InventoryRefProduct,
			ReferenceID:    &req.ProductID,
			ActorUserID:    actorUserID,
			Note:           note,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	s.handleStockSideEffects(ctx, result.Product, result.QuantityBefore, result.QuantityAfter)

	loaded, err := s.getInventoryItem(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (s *inventoryService) ListHistory(ctx context.Context, productID uint, req *dto.ListInventoryHistoryRequest) ([]dto.InventoryAdjustmentResponse, int64, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := s.db.WithContext(ctx).Model(&models.InventoryAdjustment{}).
		Where("product_id = ?", productID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	offset := (page - 1) * limit
	var rows []models.InventoryAdjustment
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Preload("Product").
		Preload("Actor").
		Find(&rows).Error; err != nil {
		return nil, 0, utils.ErrInternal(err)
	}

	return mapAdjustmentRows(rows), total, nil
}

func (s *inventoryService) ListRecentAdjustments(ctx context.Context, limit int) ([]dto.InventoryAdjustmentResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var rows []models.InventoryAdjustment
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Preload("Product").
		Preload("Actor").
		Find(&rows).Error; err != nil {
		return nil, utils.ErrInternal(err)
	}
	return mapAdjustmentRows(rows), nil
}

func (s *inventoryService) ApplyDelta(ctx context.Context, tx *gorm.DB, params DeltaParams) error {
	result, err := s.applyDeltaLocked(ctx, tx, params)
	if err != nil {
		return err
	}
	// Side effects run after the outer transaction commits.
	_ = result
	return nil
}

// RunStockSideEffects applies notification/workflow hooks after a committed stock change.
func (s *inventoryService) RunStockSideEffects(ctx context.Context, product models.Product, before, after int) {
	s.handleStockSideEffects(ctx, product, before, after)
}

func (s *inventoryService) SetAbsoluteStock(ctx context.Context, productID uint, newStock int, actorUserID *uint, note, adjustmentType string) error {
	if newStock < 0 {
		return utils.ErrBadRequest("stock cannot be negative")
	}
	if adjustmentType == "" {
		adjustmentType = constants.InventoryAdjAdminSet
	}

	var result deltaResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, productID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.ErrNotFound("product not found")
			}
			return utils.ErrInternal(err)
		}
		if !product.TrackInventory {
			product.Stock = newStock
			return tx.Save(&product).Error
		}

		delta := newStock - product.Stock
		if delta == 0 {
			result = deltaResult{Product: product, QuantityBefore: product.Stock, QuantityAfter: product.Stock}
			return nil
		}

		var err error
		result, err = s.applyDeltaLocked(ctx, tx, DeltaParams{
			ProductID:      productID,
			Delta:          delta,
			AdjustmentType: adjustmentType,
			ReferenceType:  constants.InventoryRefProduct,
			ReferenceID:    &productID,
			ActorUserID:    actorUserID,
			Note:           note,
		})
		return err
	})
	if err != nil {
		return err
	}

	if result.QuantityBefore != result.QuantityAfter {
		s.handleStockSideEffects(ctx, result.Product, result.QuantityBefore, result.QuantityAfter)
	}
	return nil
}

func (s *inventoryService) RecordInitialStock(ctx context.Context, productID uint, quantity int) error {
	if quantity <= 0 {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.First(&product, productID).Error; err != nil {
			return err
		}
		if !product.TrackInventory {
			return nil
		}
		adj := models.InventoryAdjustment{
			ProductID:      productID,
			QuantityDelta:  quantity,
			QuantityBefore: 0,
			QuantityAfter:  quantity,
			AdjustmentType: constants.InventoryAdjInitial,
			ReferenceType:  constants.InventoryRefProduct,
			ReferenceID:    &productID,
			Note:           "initial stock on create",
		}
		return tx.Create(&adj).Error
	})
}

func (s *inventoryService) DecrementForSale(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (deltaResult, error) {
	refID := orderID
	return s.applyDeltaLocked(ctx, tx, DeltaParams{
		ProductID:      productID,
		Delta:          -quantity,
		AdjustmentType: constants.InventoryAdjSale,
		ReferenceType:  constants.InventoryRefOrder,
		ReferenceID:    &refID,
		Note:           fmt.Sprintf("sale for order %d", orderID),
	})
}

func (s *inventoryService) RestoreForOrderCancel(ctx context.Context, tx *gorm.DB, orderID uint, productID uint, quantity int) (deltaResult, error) {
	var product models.Product
	if err := tx.First(&product, productID).Error; err != nil {
		return deltaResult{}, err
	}
	if !shouldDecrementProductStock(product) {
		return deltaResult{Product: product, QuantityBefore: product.Stock, QuantityAfter: product.Stock}, nil
	}

	refID := orderID
	return s.applyDeltaLocked(ctx, tx, DeltaParams{
		ProductID:             productID,
		Delta:                 quantity,
		AdjustmentType:        constants.InventoryAdjOrderCancel,
		ReferenceType:         constants.InventoryRefOrder,
		ReferenceID:           &refID,
		Note:                  fmt.Sprintf("order %d cancelled", orderID),
		SkipAvailabilityCheck: true,
	})
}

func (s *inventoryService) applyDeltaLocked(ctx context.Context, tx *gorm.DB, params DeltaParams) (deltaResult, error) {
	var product models.Product
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, params.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return deltaResult{}, utils.ErrNotFound("product not found")
		}
		return deltaResult{}, utils.ErrInternal(err)
	}

	before := product.Stock
	if !product.TrackInventory {
		return deltaResult{Product: product, QuantityBefore: before, QuantityAfter: before}, nil
	}

	after := before + params.Delta
	if after < 0 {
		return deltaResult{}, utils.ErrBadRequest(fmt.Sprintf("insufficient stock for product: %s", product.Name))
	}
	if params.Delta < 0 && !params.SkipAvailabilityCheck && !product.AllowBackorder {
		if !isProductStockAvailable(product, -params.Delta) {
			return deltaResult{}, utils.ErrBadRequest(fmt.Sprintf("insufficient stock for product: %s", product.Name))
		}
	}

	product.Stock = after
	if err := tx.Save(&product).Error; err != nil {
		return deltaResult{}, utils.ErrInternal(err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"sku":              product.SKU,
		"track_inventory":  product.TrackInventory,
		"allow_backorder":  product.AllowBackorder,
	})

	adj := models.InventoryAdjustment{
		ProductID:      product.ID,
		QuantityDelta:  params.Delta,
		QuantityBefore: before,
		QuantityAfter:  after,
		AdjustmentType: params.AdjustmentType,
		ReferenceType:  params.ReferenceType,
		ReferenceID:    params.ReferenceID,
		ActorUserID:    params.ActorUserID,
		Note:           params.Note,
		Metadata:       datatypes.JSON(meta),
	}
	if err := tx.Create(&adj).Error; err != nil {
		return deltaResult{}, utils.ErrInternal(err)
	}

	_ = ctx
	return deltaResult{Product: product, QuantityBefore: before, QuantityAfter: after}, nil
}

func (s *inventoryService) handleStockSideEffects(ctx context.Context, product models.Product, before, after int) {
	if !product.TrackInventory || before == after {
		return
	}

	if before == 0 && after > 0 && s.notifier != nil {
		_ = s.notifier.NotifyBackInStock(product.ID, product.Name, product.Slug)
		if s.engine != nil {
			_ = applyWorkflowEvent(ctx, s.engine, workflow.TransitionRequest{
				WorkflowKey: constants.WorkflowEntityProduct,
				EntityID:    product.ID,
				Event:       "restock",
				ActorRole:   constants.RoleAdmin,
			})
		}
	}

	if after == 0 && before > 0 && s.engine != nil {
		_ = applyWorkflowEvent(ctx, s.engine, workflow.TransitionRequest{
			WorkflowKey: constants.WorkflowEntityProduct,
			EntityID:    product.ID,
			Event:       "mark_out_of_stock",
			ActorRole:   constants.RoleAdmin,
		})
	}
}

func (s *inventoryService) getInventoryItem(ctx context.Context, productID uint) (*dto.InventoryItemResponse, error) {
	var row inventoryListRow
	err := s.db.WithContext(ctx).Model(&models.Product{}).
		Select(`
			products.*,
			COALESCE(waitlist.waitlist_count, 0) AS waitlist_count,
			COALESCE(sales.units_sold_30d, 0) AS units_sold_30d`).
		Joins(`LEFT JOIN (
			SELECT product_id, COUNT(*) AS waitlist_count
			FROM stock_notifications
			WHERE status = ?
			GROUP BY product_id
		) waitlist ON waitlist.product_id = products.id`, constants.StockNotificationStatusActive).
		Joins(`LEFT JOIN (
			SELECT oi.product_id, COALESCE(SUM(oi.quantity), 0) AS units_sold_30d
			FROM order_items oi
			INNER JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL
			WHERE o.created_at >= ? AND o.status IN ?
			GROUP BY oi.product_id
		) sales ON sales.product_id = products.id`,
			time.Now().AddDate(0, 0, -30),
			revenueOrderStatuses,
		).
		Where("products.id = ?", productID).
		Preload("WorkflowState").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	item := toInventoryItemResponse(&row.Product, row.WaitlistCount, row.UnitsSold30d)
	return &item, nil
}

func toInventoryItemResponse(p *models.Product, waitlistCount, unitsSold30d int64) dto.InventoryItemResponse {
	imageURL := ""
	if len(p.Images) > 0 {
		imageURL = p.Images[0]
	}

	item := dto.InventoryItemResponse{
		ID:                p.ID,
		Name:              p.Name,
		SKU:               p.SKU,
		Slug:              p.Slug,
		ImageURL:          imageURL,
		Stock:             p.Stock,
		LowStockThreshold: p.LowStockThreshold,
		TrackInventory:    p.TrackInventory,
		AllowBackorder:    p.AllowBackorder,
		WarehouseLocation: p.WarehouseLocation,
		Status:            p.Status,
		WaitlistCount:     waitlistCount,
		UnitsSold30d:      unitsSold30d,
		StockStatus:       computeStockStatus(p),
		UpdatedAt:         p.UpdatedAt,
	}
	if p.WorkflowState != nil {
		item.WorkflowState = dto.ToStateView(p.WorkflowState)
	}
	return item
}

func computeStockStatus(p *models.Product) string {
	if !p.TrackInventory {
		return constants.InventoryStockNotTracked
	}
	if p.Stock == 0 {
		return constants.InventoryStockOut
	}
	if p.Stock <= p.LowStockThreshold {
		return constants.InventoryStockLow
	}
	return constants.InventoryStockHealthy
}

func mapReasonToAdjustmentType(reason string) string {
	switch reason {
	case "receive":
		return constants.InventoryAdjReceive
	case "damage":
		return constants.InventoryAdjDamage
	case "correction", "cycle_count", "shrinkage", "other":
		return constants.InventoryAdjCorrection
	default:
		return constants.InventoryAdjAdminDelta
	}
}

func mapAdjustmentRows(rows []models.InventoryAdjustment) []dto.InventoryAdjustmentResponse {
	out := make([]dto.InventoryAdjustmentResponse, 0, len(rows))
	for _, row := range rows {
		resp := dto.InventoryAdjustmentResponse{
			ID:             row.ID,
			ProductID:      row.ProductID,
			QuantityDelta:  row.QuantityDelta,
			QuantityBefore: row.QuantityBefore,
			QuantityAfter:  row.QuantityAfter,
			AdjustmentType: row.AdjustmentType,
			ReferenceType:  row.ReferenceType,
			ReferenceID:    row.ReferenceID,
			Note:           row.Note,
			CreatedAt:      row.CreatedAt,
		}
		if row.Product != nil {
			resp.ProductName = row.Product.Name
			resp.ProductSKU = row.Product.SKU
		}
		if row.Actor != nil {
			resp.ActorName = strings.TrimSpace(row.Actor.FirstName + " " + row.Actor.LastName)
		}
		out = append(out, resp)
	}
	return out
}

func (s *inventoryService) BulkAdjustStock(
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

		var product models.Product
		if err := s.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error; err != nil {
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

func (s *inventoryService) SendLowStockAlerts(ctx context.Context) error {
	var products []models.Product
	err := s.db.WithContext(ctx).
		Where("track_inventory = ? AND status = ? AND stock <= low_stock_threshold", true, constants.ProductStatusActive).
		Order("stock ASC").
		Find(&products).Error
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
			"low_stock_count":   len(lowStock),
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

func (s *inventoryService) resolveAlertRecipients(ctx context.Context) (alertRecipients, error) {
	var admins []models.User
	if err := s.db.WithContext(ctx).
		Where("role = ?", constants.RoleAdmin).
		Find(&admins).Error; err != nil {
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

func (s *inventoryService) RestockForReturn(ctx context.Context, returnID uint) error {
	var ret models.Return
	if err := s.db.WithContext(ctx).Preload("Order.Items").First(&ret, returnID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("return not found")
		}
		return utils.ErrInternal(err)
	}
	if ret.Order == nil || len(ret.Order.Items) == 0 {
		return nil
	}

	var changes []deltaResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range ret.Order.Items {
			var product models.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return err
			}
			if !shouldDecrementProductStock(product) {
				continue
			}

			var existing int64
			if err := tx.Model(&models.InventoryAdjustment{}).
				Where("reference_type = ? AND reference_id = ? AND product_id = ? AND adjustment_type = ?",
					constants.InventoryRefReturn, returnID, item.ProductID, constants.InventoryAdjReturnRestock).
				Count(&existing).Error; err != nil {
				return err
			}
			if existing > 0 {
				continue
			}

			refID := returnID
			change, err := s.applyDeltaLocked(ctx, tx, DeltaParams{
				ProductID:             item.ProductID,
				Delta:                 item.Quantity,
				AdjustmentType:        constants.InventoryAdjReturnRestock,
				ReferenceType:         constants.InventoryRefReturn,
				ReferenceID:           &refID,
				Note:                  fmt.Sprintf("return %d item received", returnID),
				SkipAvailabilityCheck: true,
			})
			if err != nil {
				return err
			}
			if change.QuantityBefore != change.QuantityAfter {
				changes = append(changes, change)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, change := range changes {
		s.handleStockSideEffects(ctx, change.Product, change.QuantityBefore, change.QuantityAfter)
	}
	return nil
}
