package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	domaininventory "github.com/alireza-akbarzadeh/luxe/internal/domain/inventory"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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

// DeltaResult captures stock change outcome for side effects.
type DeltaResult struct {
	Product        models.Product
	QuantityBefore int
	QuantityAfter  int
}

// Commands orchestrates inventory write use cases.
type Commands struct {
	repo *postgres.InventoryRepository
}

// NewCommands creates inventory command use cases.
func NewCommands(repo *postgres.InventoryRepository) *Commands {
	return &Commands{repo: repo}
}

// ApplyDeltaLocked applies a stock delta under row lock within a transaction.
func (c *Commands) ApplyDeltaLocked(ctx context.Context, tx *gorm.DB, params DeltaParams) (DeltaResult, error) {
	product, err := c.repo.GetProductForUpdate(ctx, tx, params.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DeltaResult{}, utils.ErrNotFound("product not found")
		}
		return DeltaResult{}, utils.ErrInternal(err)
	}

	before := product.Stock
	stock := productStockFromModel(*product)
	if !stock.TrackInventory {
		return DeltaResult{Product: *product, QuantityBefore: before, QuantityAfter: before}, nil
	}

	if err := domaininventory.CanApplyDelta(stock, before, params.Delta, params.SkipAvailabilityCheck); err != nil {
		switch {
		case errors.Is(err, domaininventory.ErrInsufficientStock),
			errors.Is(err, domaininventory.ErrNegativeStock):
			return DeltaResult{}, utils.ErrBadRequest(fmt.Sprintf("insufficient stock for product: %s", product.Name))
		default:
			return DeltaResult{}, utils.ErrInternal(err)
		}
	}

	after := before + params.Delta

	product.Stock = after
	if err := c.repo.SaveProduct(ctx, tx, product); err != nil {
		return DeltaResult{}, utils.ErrInternal(err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"sku":             product.SKU,
		"track_inventory": product.TrackInventory,
		"allow_backorder": product.AllowBackorder,
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
	if err := c.repo.CreateAdjustment(ctx, tx, &adj); err != nil {
		return DeltaResult{}, utils.ErrInternal(err)
	}

	return DeltaResult{Product: *product, QuantityBefore: before, QuantityAfter: after}, nil
}

// AdjustStock applies a manual stock adjustment in its own transaction.
func (c *Commands) AdjustStock(ctx context.Context, actorUserID *uint, req *dto.AdjustInventoryRequest) (DeltaResult, error) {
	if req.Delta == 0 {
		return DeltaResult{}, utils.ErrBadRequest("adjustment delta cannot be zero")
	}

	adjType := mapReasonToAdjustmentType(req.Reason)
	note := strings.TrimSpace(req.Note)
	if note == "" {
		note = req.Reason
	}

	var result DeltaResult
	err := c.repo.Transaction(ctx, func(tx *gorm.DB) error {
		var err error
		result, err = c.ApplyDeltaLocked(ctx, tx, DeltaParams{
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
	return result, err
}

// SetAbsoluteStock sets stock to an absolute value.
func (c *Commands) SetAbsoluteStock(ctx context.Context, productID uint, newStock int, actorUserID *uint, note, adjustmentType string) (DeltaResult, error) {
	if newStock < 0 {
		return DeltaResult{}, utils.ErrBadRequest("stock cannot be negative")
	}
	if adjustmentType == "" {
		adjustmentType = constants.InventoryAdjAdminSet
	}

	var result DeltaResult
	err := c.repo.Transaction(ctx, func(tx *gorm.DB) error {
		product, err := c.repo.GetProductForUpdate(ctx, tx, productID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.ErrNotFound("product not found")
			}
			return utils.ErrInternal(err)
		}
		if !product.TrackInventory {
			product.Stock = newStock
			return c.repo.SaveProduct(ctx, tx, product)
		}

		delta := newStock - product.Stock
		if delta == 0 {
			result = DeltaResult{Product: *product, QuantityBefore: product.Stock, QuantityAfter: product.Stock}
			return nil
		}

		result, err = c.ApplyDeltaLocked(ctx, tx, DeltaParams{
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
	return result, err
}

// RecordInitialStock records the opening ledger entry for a new product.
func (c *Commands) RecordInitialStock(ctx context.Context, productID uint, quantity int) error {
	if quantity <= 0 {
		return nil
	}
	return c.repo.Transaction(ctx, func(tx *gorm.DB) error {
		product, err := c.repo.GetProduct(ctx, tx, productID)
		if err != nil {
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
		return c.repo.CreateAdjustment(ctx, tx, &adj)
	})
}

// DecrementForSale reduces stock for an order line.
func (c *Commands) DecrementForSale(ctx context.Context, tx *gorm.DB, orderID, productID uint, quantity int) (DeltaResult, error) {
	refID := orderID
	return c.ApplyDeltaLocked(ctx, tx, DeltaParams{
		ProductID:      productID,
		Delta:          -quantity,
		AdjustmentType: constants.InventoryAdjSale,
		ReferenceType:  constants.InventoryRefOrder,
		ReferenceID:    &refID,
		Note:           fmt.Sprintf("sale for order %d", orderID),
	})
}

// RestoreForOrderCancel restores stock when an order is cancelled.
func (c *Commands) RestoreForOrderCancel(ctx context.Context, tx *gorm.DB, orderID, productID uint, quantity int) (DeltaResult, error) {
	product, err := c.repo.GetProduct(ctx, tx, productID)
	if err != nil {
		return DeltaResult{}, err
	}
	if !domaininventory.ShouldAdjustStock(productStockFromModel(*product)) {
		return DeltaResult{Product: *product, QuantityBefore: product.Stock, QuantityAfter: product.Stock}, nil
	}

	refID := orderID
	return c.ApplyDeltaLocked(ctx, tx, DeltaParams{
		ProductID:             productID,
		Delta:                 quantity,
		AdjustmentType:        constants.InventoryAdjOrderCancel,
		ReferenceType:         constants.InventoryRefOrder,
		ReferenceID:           &refID,
		Note:                  fmt.Sprintf("order %d cancelled", orderID),
		SkipAvailabilityCheck: true,
	})
}

// RestockForReturn restocks items when a return is received.
func (c *Commands) RestockForReturn(ctx context.Context, returnID uint) ([]DeltaResult, error) {
	ret, err := c.repo.GetReturnWithOrder(ctx, returnID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("return not found")
		}
		return nil, utils.ErrInternal(err)
	}
	if ret.Order == nil || len(ret.Order.Items) == 0 {
		return nil, nil
	}

	var changes []DeltaResult
	err = c.repo.Transaction(ctx, func(tx *gorm.DB) error {
		for _, item := range ret.Order.Items {
			product, err := c.repo.GetProduct(ctx, tx, item.ProductID)
			if err != nil {
				return err
			}
			if !domaininventory.ShouldAdjustStock(productStockFromModel(*product)) {
				continue
			}

			existing, err := c.repo.CountReturnRestockAdjustments(ctx, tx, returnID, item.ProductID)
			if err != nil {
				return err
			}
			if existing > 0 {
				continue
			}

			refID := returnID
			change, err := c.ApplyDeltaLocked(ctx, tx, DeltaParams{
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
	return changes, err
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

func productStockFromModel(product models.Product) domaininventory.ProductStock {
	return domaininventory.ProductStock{
		TrackInventory: product.TrackInventory,
		AllowBackorder: product.AllowBackorder,
		Stock:          product.Stock,
	}
}
