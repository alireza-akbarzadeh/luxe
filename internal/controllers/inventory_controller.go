package controllers

import (
	"errors"
	"fmt"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type InventoryController struct {
	inventoryService services.InventoryServiceInterface
	validate         *validator.Validate
}

func NewInventoryController(inventoryService services.InventoryServiceInterface) *InventoryController {
	return &InventoryController{
		inventoryService: inventoryService,
		validate:         validator.New(),
	}
}

// GetOverview godoc
// @Summary      Inventory overview KPIs
// @Description  Returns low-stock, out-of-stock, and waitlist totals for the admin inventory dashboard
// @Tags         inventory
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.InventoryOverviewResponse}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/inventory/overview [get]
func (ctrl *InventoryController) GetOverview(c *gin.Context) {
	overview, err := ctrl.inventoryService.GetOverview(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load inventory overview")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, overview)
}

// List godoc
// @Summary      List inventory items
// @Description  Paginated product inventory rows with stock status, waitlist counts, and 30-day velocity
// @Tags         inventory
// @Produce      json
// @Security     BearerAuth
// @Param        page          query int    false "Page number" default(1)
// @Param        limit         query int    false "Page size" default(20)
// @Param        search        query string false "Search by name or SKU"
// @Param        stock_status  query string false "Filter" Enums(all,low,out,healthy,not_tracked)
// @Param        sort          query string false "Sort" Enums(stock_asc,stock_desc,name_asc,waitlist_desc,velocity_desc)
// @Success      200 {object} utils.Response{data=dto.InventoryListData}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/inventory [get]
func (ctrl *InventoryController) List(c *gin.Context) {
	var req dto.ListInventoryRequest
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}
	if req.StockStatus == "" {
		req.StockStatus = constants.InventoryStockAll
	}

	items, total, err := ctrl.inventoryService.List(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list inventory")
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.InventoryListData{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// Adjust godoc
// @Summary      Adjust product stock
// @Description  Apply a signed stock delta with reason (receive, correction, damage, etc.)
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.AdjustInventoryRequest true "Adjustment payload"
// @Success      200 {object} utils.Response{data=dto.InventoryItemResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/inventory/adjust [post]
func (ctrl *InventoryController) Adjust(c *gin.Context) {
	var req dto.AdjustInventoryRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	var actorID *uint
	if userID, ok := middleware.GetUserID(c); ok {
		actorID = &userID
	}

	item, err := ctrl.inventoryService.AdjustStock(c.Request.Context(), actorID, &req)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			utils.NotFoundResponse(c, "product not found")
			return
		}
		utils.HandleServiceError(c, err, "failed to adjust stock")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, item)
}

// ListHistory godoc
// @Summary      Product inventory history
// @Description  Paginated adjustment ledger for a single product
// @Tags         inventory
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  int true "Product ID"
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Page size" default(20)
// @Success      200 {object} utils.Response{data=dto.InventoryHistoryData}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/inventory/products/{id}/history [get]
func (ctrl *InventoryController) ListHistory(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.ListInventoryHistoryRequest
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	rows, total, err := ctrl.inventoryService.ListHistory(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load inventory history")
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 20
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.InventoryHistoryData{
		Adjustments: rows,
		Total:       total,
		Page:        page,
		Limit:       limit,
	})
}

// ListRecent godoc
// @Summary      Recent inventory adjustments
// @Description  Latest stock movements across all products
// @Tags         inventory
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max rows" default(10)
// @Success      200 {object} utils.Response{data=[]dto.InventoryAdjustmentResponse}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/inventory/adjustments/recent [get]
func (ctrl *InventoryController) ListRecent(c *gin.Context) {
	limit := 10
	if l, ok := parseOptionalIntQuery(c, "limit"); ok && l > 0 {
		limit = l
	}

	rows, err := ctrl.inventoryService.ListRecentAdjustments(c.Request.Context(), limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load recent adjustments")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, rows)
}

func parseOptionalIntQuery(c *gin.Context, key string) (int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return 0, false
	}
	var value int
	if _, err := fmt.Sscanf(raw, "%d", &value); err != nil {
		return 0, false
	}
	return value, true
}
