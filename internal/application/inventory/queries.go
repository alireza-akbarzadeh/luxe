package inventory

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
)

// Queries orchestrates inventory read use cases.
type Queries struct {
	repo             *postgres.InventoryRepository
	revenueStatuses  []string
}

// NewQueries creates inventory query use cases.
func NewQueries(repo *postgres.InventoryRepository, revenueStatuses []string) *Queries {
	return &Queries{repo: repo, revenueStatuses: revenueStatuses}
}

// GetOverview returns aggregate inventory counters.
func (q *Queries) GetOverview(ctx context.Context) (*dto.InventoryOverviewResponse, error) {
	overview := &dto.InventoryOverviewResponse{}

	low, err := q.repo.CountLowStockActive(ctx)
	if err != nil {
		return nil, err
	}
	overview.LowStockCount = low

	out, err := q.repo.CountOutOfStockActive(ctx)
	if err != nil {
		return nil, err
	}
	overview.OutOfStockCount = out

	notTracked, err := q.repo.CountNotTracked(ctx)
	if err != nil {
		return nil, err
	}
	overview.NotTrackedCount = notTracked

	tracked, err := q.repo.CountTrackedSKUs(ctx)
	if err != nil {
		return nil, err
	}
	overview.TrackedSKUCount = tracked

	units, err := q.repo.SumTrackedUnits(ctx)
	if err != nil {
		return nil, err
	}
	overview.TotalUnitsOnHand = units

	waitlist, err := q.repo.CountActiveWaitlist(ctx)
	if err != nil {
		return nil, err
	}
	overview.WaitlistTotal = waitlist

	return overview, nil
}

// List returns paginated inventory items.
func (q *Queries) List(ctx context.Context, req *dto.ListInventoryRequest) ([]dto.InventoryItemResponse, int64, error) {
	rows, total, err := q.repo.ListInventoryItems(ctx, req, q.revenueStatuses)
	if err != nil {
		return nil, 0, err
	}
	items := make([]dto.InventoryItemResponse, 0, len(rows))
	for i := range rows {
		items = append(items, ToInventoryItemResponse(&rows[i].Product, rows[i].WaitlistCount, rows[i].UnitsSold30d))
	}
	return items, total, nil
}

// GetItem loads a single inventory item view.
func (q *Queries) GetItem(ctx context.Context, productID uint) (*dto.InventoryItemResponse, error) {
	row, err := q.repo.GetInventoryItem(ctx, productID, q.revenueStatuses)
	if err != nil {
		return nil, err
	}
	item := ToInventoryItemResponse(&row.Product, row.WaitlistCount, row.UnitsSold30d)
	return &item, nil
}

// ListHistory returns paginated adjustments for a product.
func (q *Queries) ListHistory(ctx context.Context, productID uint, req *dto.ListInventoryHistoryRequest) ([]dto.InventoryAdjustmentResponse, int64, error) {
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
	offset := (page - 1) * limit

	total, err := q.repo.CountAdjustments(ctx, productID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.repo.ListAdjustments(ctx, productID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	return MapAdjustmentRows(rows), total, nil
}

// ListRecentAdjustments returns the latest global adjustments.
func (q *Queries) ListRecentAdjustments(ctx context.Context, limit int) ([]dto.InventoryAdjustmentResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	rows, err := q.repo.ListRecentAdjustments(ctx, limit)
	if err != nil {
		return nil, err
	}
	return MapAdjustmentRows(rows), nil
}

// FindProductBySKU loads a product by SKU.
func (q *Queries) FindProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	return q.repo.FindProductBySKU(ctx, sku)
}

// FindLowStockTracked returns tracked products at or below threshold.
func (q *Queries) FindLowStockTracked(ctx context.Context) ([]models.Product, error) {
	return q.repo.FindLowStockTracked(ctx)
}

// FindAdmins loads admin users for alert notifications.
func (q *Queries) FindAdmins(ctx context.Context) ([]models.User, error) {
	return q.repo.FindAdmins(ctx)
}

// ToInventoryItemResponse maps a product to an inventory list item.
func ToInventoryItemResponse(p *models.Product, waitlistCount, unitsSold30d int64) dto.InventoryItemResponse {
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
		StockStatus:       ComputeStockStatus(p),
		UpdatedAt:         p.UpdatedAt,
	}
	if p.WorkflowState != nil {
		item.WorkflowState = dto.ToStateView(p.WorkflowState)
	}
	return item
}

// ComputeStockStatus derives the inventory health label.
func ComputeStockStatus(p *models.Product) string {
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

// MapAdjustmentRows maps adjustment models to API responses.
func MapAdjustmentRows(rows []models.InventoryAdjustment) []dto.InventoryAdjustmentResponse {
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
