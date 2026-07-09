package handlers

import (
	"fmt"
	"net/http"
	"time"

	appadmin "github.com/alireza-akbarzadeh/luxe/internal/application/admin"
	orderfacade "github.com/alireza-akbarzadeh/luxe/internal/application/order/facade"
	appwebhook "github.com/alireza-akbarzadeh/luxe/internal/application/webhookevent"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AdminHandler struct {
	adminService   *appadmin.Service
	orderService   *orderfacade.Service
	webhookQueries *appwebhook.Queries
	validate       *validator.Validate
}

func NewAdminHandler(svc *appadmin.Service, orderSvc *orderfacade.Service, webhookQueries *appwebhook.Queries) *AdminHandler {
	return &AdminHandler{adminService: svc, orderService: orderSvc, webhookQueries: webhookQueries, validate: validator.New()}
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
func (ctrl *AdminHandler) GetStats(c *gin.Context) {
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
func (ctrl *AdminHandler) GetDashboardOverview(c *gin.Context) {
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

// GetDashboardHealth returns platform health metrics for the admin dashboard.
// @Summary      Admin dashboard health
// @Description  Returns database latency, webhook error rate, and queue depth indicators.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminDashboardHealth}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/dashboard/health [get]
func (ctrl *AdminHandler) GetDashboardHealth(c *gin.Context) {
	health, err := ctrl.adminService.GetDashboardHealth(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get dashboard health")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, health)
}

// ExportDashboardCSV streams a CSV export of dashboard revenue series.
// @Summary      Export dashboard report (admin)
// @Description  Returns a CSV file of daily revenue, orders, and AOV for the selected period.
// @Tags         Admin
// @Produce      text/csv
// @Security     BearerAuth
// @Param        period  query  string  false  "Period: 7d, 30d, or 90d (default 30d)"
// @Param        format  query  string  false  "Format: csv (default csv)"
// @Success      200  {file}   binary
// @Failure      401  {object} utils.Response
// @Failure      403  {object} utils.Response
// @Failure      500  {object} utils.Response
// @Router       /admin/dashboard/export [get]
func (ctrl *AdminHandler) ExportDashboardCSV(c *gin.Context) {
	var filters dto.AdminDashboardExportFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	if filters.Period == "" {
		filters.Period = "30d"
	}
	if filters.Format == "" {
		filters.Format = "csv"
	}

	data, err := ctrl.adminService.ExportDashboardCSV(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to export dashboard report")
		return
	}
	filename := fmt.Sprintf("dashboard_%s_%s.csv", filters.Period, time.Now().UTC().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// GetNavPreferences returns favorites and recent pages for the current admin user.
// @Summary      Admin nav preferences
// @Description  Returns persisted favorites and recently visited admin pages for the authenticated user.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminNavPreferencesResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/nav/preferences [get]
func (ctrl *AdminHandler) GetNavPreferences(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	prefs, err := ctrl.adminService.GetNavPreferences(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get nav preferences")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, prefs)
}

// UpdateNavPreferences saves favorites and recent pages for the current admin user.
// @Summary      Update admin nav preferences
// @Description  Upserts favorites and recently visited admin pages for the authenticated user.
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateAdminNavPreferencesRequest true "Nav preferences"
// @Success      200 {object} utils.Response{data=dto.AdminNavPreferencesResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/nav/preferences [put]
func (ctrl *AdminHandler) UpdateNavPreferences(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req dto.UpdateAdminNavPreferencesRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	prefs, err := ctrl.adminService.UpdateNavPreferences(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update nav preferences")
		return
	}
	utils.SuccessResponse(c, "nav preferences updated", prefs)
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
func (ctrl *AdminHandler) GetRevenueReport(c *gin.Context) {
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
func (ctrl *AdminHandler) GetSalesFeedSnapshot(c *gin.Context) {
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
// @Param        role              query  string  false  "Filter by role (admin|user)"
// @Param        is_active         query  bool    false  "Filter by active status"
// @Param        membership_tier   query  string  false  "Filter by membership tier (free|plus)"
// @Param        customer_segment  query  string  false  "Filter by CRM segment (vip|loyal|new|at_risk)"
// @Success      200 {object} utils.Response{data=object{users=[]dto.AdminUserResponse,total=int,limit=int,offset=int}}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users [get]
func (ctrl *AdminHandler) ListUsers(c *gin.Context) {
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
func (ctrl *AdminHandler) UpdateUserRole(c *gin.Context) {
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
func (ctrl *AdminHandler) ToggleUserActive(c *gin.Context) {
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

// GetCustomerDetail returns a full customer profile (admin only).
// @Summary      Get customer detail (admin)
// @Description  Returns customer profile with purchase stats, loyalty tier, and CRM fields.
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "User ID"
// @Success      200 {object} utils.Response{data=dto.AdminCustomerDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id} [get]
func (ctrl *AdminHandler) GetCustomerDetail(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	detail, err := ctrl.adminService.GetCustomerDetail(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get customer detail")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, detail)
}

// ListCustomerAddresses returns saved addresses for a customer (admin only).
// @Summary      List customer addresses (admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "User ID"
// @Success      200 {object} utils.Response{data=object{addresses=[]models.Address}}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id}/addresses [get]
func (ctrl *AdminHandler) ListCustomerAddresses(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	addresses, err := ctrl.adminService.ListCustomerAddresses(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list customer addresses")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{"addresses": addresses})
}

// UpdateCustomerNotes updates admin CRM notes on a customer (admin only).
// @Summary      Update customer admin notes (admin)
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path  int                              true  "User ID"
// @Param        request body  dto.UpdateCustomerNotesRequest   true  "Admin notes"
// @Success      200 {object} utils.Response{data=dto.AdminCustomerDetailResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id}/notes [patch]
func (ctrl *AdminHandler) UpdateCustomerNotes(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCustomerNotesRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	detail, err := ctrl.adminService.UpdateCustomerNotes(c.Request.Context(), userID, req.AdminNotes)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update customer notes")
		return
	}
	utils.SuccessResponse(c, "notes updated", detail)
}

// UpdateCustomerSegment assigns a CRM segment to a customer (admin only).
// @Summary      Update customer segment (admin)
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path  int                                true  "User ID"
// @Param        request body  dto.UpdateCustomerSegmentRequest   true  "CRM segment"
// @Success      200 {object} utils.Response{data=dto.AdminCustomerDetailResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/users/{id}/segment [patch]
func (ctrl *AdminHandler) UpdateCustomerSegment(c *gin.Context) {
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateCustomerSegmentRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	detail, err := ctrl.adminService.UpdateCustomerSegment(c.Request.Context(), userID, req.CustomerSegment)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update customer segment")
		return
	}
	utils.SuccessResponse(c, "segment updated", detail)
}

// GetCustomerStats returns aggregate CRM metrics (admin only).
// @Summary      Customer analytics (admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminCustomerStats}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/customers/stats [get]
func (ctrl *AdminHandler) GetCustomerStats(c *gin.Context) {
	stats, err := ctrl.adminService.GetCustomerStats(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get customer stats")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, stats)
}

// BulkUpdateOrderStatus updates the status of multiple orders atomically (admin only).
// Deprecated: prefer POST /workflows/order/{id}/transition per order so lifecycle hooks and audit run correctly.
// @Summary      Bulk update order status (admin) [deprecated]
// @Description  Deprecated — use workflow transitions on order detail instead. Applies the given status to all specified order IDs. Max 500 IDs per call.
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
func (ctrl *AdminHandler) BulkUpdateOrderStatus(c *gin.Context) {
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
func (ctrl *AdminHandler) ExportOrdersCSV(c *gin.Context) {
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

// ExportProductsCSV streams a CSV file of filtered products (admin only).
// @Summary      Export products CSV (admin)
// @Description  Returns a CSV file of products filtered by status, name, SKU, category, brand, price, and type (max 10 000 rows).
// @Tags         Admin
// @Produce      text/csv
// @Security     BearerAuth
// @Param        status      query  string   false  "Product status (active|draft|archived)"
// @Param        name        query  string   false  "Filter by product name"
// @Param        sku         query  string   false  "Filter by SKU"
// @Param        category_id query  int      false  "Filter by category ID"
// @Param        brand_id    query  int      false  "Filter by brand ID"
// @Param        min_price   query  number   false  "Minimum price"
// @Param        max_price   query  number   false  "Maximum price"
// @Param        is_digital  query  bool     false  "Digital products only"
// @Param        format      query  string   false  "Format: csv (default csv)"
// @Success      200  {file}   binary
// @Failure      401  {object} utils.Response
// @Failure      403  {object} utils.Response
// @Failure      500  {object} utils.Response
// @Router       /admin/products/export [get]
func (ctrl *AdminHandler) ExportProductsCSV(c *gin.Context) {
	var filters dto.AdminProductExportFilters
	if !utils.BindAndValidateQuery(c, &filters, ctrl.validate) {
		return
	}
	if filters.Format == "" {
		filters.Format = "csv"
	}
	data, err := ctrl.adminService.ExportProductsCSV(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to export products")
		return
	}
	filename := fmt.Sprintf("products_%s.csv", time.Now().UTC().Format("20060102_150405"))
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
func (ctrl *AdminHandler) ListWebhookEvents(c *gin.Context) {
	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters := appwebhook.EventFilters{
		Source:    c.Query("source"),
		EventType: c.Query("event_type"),
		Status:    c.Query("status"),
		Limit:     limit,
		Offset:    offset,
	}
	events, total, err := ctrl.webhookQueries.List(c.Request.Context(), filters)
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
