package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AdminController struct {
	adminService        services.AdminServiceInterface
	orderService        services.OrderServiceInterface
	webhookEventService services.WebhookEventServiceInterface
	validate            *validator.Validate
}

func NewAdminController(svc services.AdminServiceInterface, orderSvc services.OrderServiceInterface, webhookSvc services.WebhookEventServiceInterface) *AdminController {
	return &AdminController{adminService: svc, orderService: orderSvc, webhookEventService: webhookSvc, validate: validator.New()}
}

// GetStats returns platform-wide statistics (admin only).
// @Summary      Platform stats (admin)
// @Description  Returns counts of users, orders, products, revenue, wallet balances, and low-stock alerts.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminStatsResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/stats [get]
func (ctrl *AdminController) GetStats(c *gin.Context) {
	stats, err := ctrl.adminService.GetStats(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get stats")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, stats)
}

// GetDashboardOverview returns KPIs, charts, and operational lists for the admin home dashboard.
// @Summary      Admin dashboard overview
// @Description  Returns period KPIs, revenue series, order status breakdown, recent orders, top products, and low-stock alerts.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        period  query  string  false  "Period: 7d, 30d, or 90d (default 30d)"
// @Success      200 {object} utils.Response{data=dto.AdminDashboardOverviewResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/dashboard/overview [get]
func (ctrl *AdminController) GetDashboardOverview(c *gin.Context) {
	var filters dto.AdminDashboardFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	if filters.Period == "" {
		filters.Period = "30d"
	}

	overview, err := ctrl.adminService.GetDashboardOverview(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get dashboard overview")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, overview)
}

// GetRevenueReport returns a daily revenue breakdown for the admin reports page.
// @Summary      Daily revenue report (admin)
// @Description  Returns period summary KPIs and a day-by-day revenue, orders, and AOV breakdown.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        period  query  string  false  "Period: 7d, 30d, or 90d (default 30d)"
// @Success      200 {object} utils.Response{data=dto.AdminRevenueReportResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/reports/revenue [get]
func (ctrl *AdminController) GetRevenueReport(c *gin.Context) {
	var filters dto.AdminRevenueReportFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	if filters.Period == "" {
		filters.Period = "30d"
	}

	report, err := ctrl.adminService.GetRevenueReport(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get revenue report")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, report)
}

// GetSalesFeedSnapshot returns today's sales metrics and recent activity for the live feed page.
// @Summary      Live sales feed snapshot
// @Description  Returns today's order totals, status breakdown, revenue series, and recent feed events.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminSalesFeedSnapshotResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/sales-feed/snapshot [get]
func (ctrl *AdminController) GetSalesFeedSnapshot(c *gin.Context) {
	snapshot, err := ctrl.adminService.GetSalesFeedSnapshot(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get sales feed snapshot")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, snapshot)
}

// ListUsers returns paginated user list with optional filters (admin only).
// @Summary      List users (admin)
// @Description  Returns paginated users with optional filters by search term, email, role, and active status.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        limit     query  int     false  "Items per page (default 20)"
// @Param        offset    query  int     false  "Offset"
// @Param        search    query  string  false  "Search name or email (partial)"
// @Param        email     query  string  false  "Email search (partial)"
// @Param        role      query  string  false  "Filter by role (admin|user)"
// @Param        is_active query  bool    false  "Filter by active status"
// @Success      200 {object} utils.Response{data=object{users=[]dto.AdminUserResponse,total=int,limit=int,offset=int}}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users [get]
func (ctrl *AdminController) ListUsers(c *gin.Context) {
	var filters dto.AdminUserFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters.Limit = limit
	filters.Offset = offset

	users, total, err := ctrl.adminService.ListUsers(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list users")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateUserRole changes the role of a user (admin only).
// @Summary      Update user role (admin)
// @Description  Sets the role slug of a user. The slug must exist in the roles table.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path  int                         true  "User ID"
// @Param        request body  dto.UpdateUserRoleRequest   true  "New role"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id}/role [patch]
func (ctrl *AdminController) UpdateUserRole(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateUserRoleRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.adminService.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		utils.HandleServiceError(c, err, "failed to update user role")
		return
	}
	utils.SuccessResponse(c, "user role updated", nil)
}

// ToggleUserActive enables or disables a user account (admin only).
// @Summary      Toggle user active status (admin)
// @Description  Enables or disables a user account.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path  int                            true  "User ID"
// @Param        request body  dto.ToggleUserActiveRequest    true  "Active flag"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id}/active [patch]
func (ctrl *AdminController) ToggleUserActive(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.ToggleUserActiveRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.adminService.ToggleUserActive(c.Request.Context(), userID, req.IsActive); err != nil {
		utils.HandleServiceError(c, err, "failed to toggle user active")
		return
	}
	utils.SuccessResponse(c, "user status updated", nil)
}

// BulkUpdateOrderStatus updates the status of multiple orders atomically (admin only).
// @Summary      Bulk update order status (admin)
// @Description  Applies the given status to all specified order IDs. Max 500 IDs per call.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.BulkUpdateOrderStatusRequest true "Order IDs and target status"
// @Success      200 {object} utils.Response{data=object{updated=int}}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/orders/bulk-status [post]
func (ctrl *AdminController) BulkUpdateOrderStatus(c *gin.Context) {
	var req dto.BulkUpdateOrderStatusRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	var actorID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		actorID = &uid
	}
	updated, err := ctrl.orderService.BulkUpdateOrderStatus(c.Request.Context(), req.OrderIDs, req.Status, actorID)
	if err != nil {
		utils.HandleServiceError(c, err, "bulk status update failed")
		return
	}
	utils.SuccessResponse(c, "orders updated", gin.H{"updated": updated})
}

// ExportOrdersCSV streams a CSV file of filtered orders (admin only).
// @Summary      Export orders CSV (admin)
// @Description  Returns a CSV file of orders filtered by status and date range (max 10 000 rows).
// @Tags         Admin
// @Produce      text/csv
// @Security     BearerAuth
// @Param        status    query  string  false  "Filter by status"
// @Param        from_date query  string  false  "Start date (YYYY-MM-DD)"
// @Param        to_date   query  string  false  "End date (YYYY-MM-DD)"
// @Success      200  {file}   binary
// @Failure      401  {object} utils.Response
// @Failure      403  {object} utils.Response
// @Failure      500  {object} utils.Response
// @Router       /admin/orders/export [get]
func (ctrl *AdminController) ExportOrdersCSV(c *gin.Context) {
	var filters dto.AdminOrderExportFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	data, err := ctrl.adminService.ExportOrdersCSV(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to export orders")
		return
	}
	filename := fmt.Sprintf("orders_%s.csv", time.Now().UTC().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// ListWebhookEvents returns paginated webhook delivery history (admin only).
// @Summary      List webhook events (admin)
// @Description  Returns paginated Stripe webhook events with optional filters by source, type, and status.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        limit      query  int     false  "Items per page (default 20)"
// @Param        offset     query  int     false  "Offset"
// @Param        source     query  string  false  "Filter by source (e.g. stripe)"
// @Param        event_type query  string  false  "Filter by event type"
// @Param        status     query  string  false  "Filter by status (received|processed|failed)"
// @Success      200 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/webhooks [get]
func (ctrl *AdminController) ListWebhookEvents(c *gin.Context) {
	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters := services.WebhookEventFilters{
		Source:    c.Query("source"),
		EventType: c.Query("event_type"),
		Status:    c.Query("status"),
		Limit:     limit,
		Offset:    offset,
	}
	events, total, err := ctrl.webhookEventService.List(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list webhook events")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
