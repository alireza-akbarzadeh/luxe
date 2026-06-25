package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	appshipment "github.com/alireza-akbarzadeh/luxe/internal/application/shipment"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ShipmentHandler struct {
	shipmentService *appshipment.Service
	validate        *validator.Validate
}

func NewShipmentHandler(shipmentService *appshipment.Service) *ShipmentHandler {
	return &ShipmentHandler{
		shipmentService: shipmentService,
		validate:        validator.New(),
	}
}

// CreateShipment creates a new shipment (admin only) and enqueues background processing.
// @Summary      Create a shipment
// @Description  Creates a shipment record and triggers a background job to process it (e.g., call carrier API).
// @Tags         Shipments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
//
//	@Param        request body object true "Shipment data" SchemaExample({
//	  "order_id":1,
//	  "carrier":"FedEx",
//	  "tracking_number":"123456789",
//	  "address_line1":"123 Main St",
//	  "city":"Springfield",
//	  "postal_code":"12345",
//	  "country":"USA"
//	})
//
// @Success      201 {object} utils.Response{data=models.Shipment}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/shipments [post]
func (ctrl *ShipmentHandler) CreateShipment(c *gin.Context) {
	var req appshipment.CreateShipmentRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	shipment, err := ctrl.shipmentService.CreateShipment(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create shipment")
		return
	}

	utils.CreatedResponse(c, constants.MsgCreateSuccess, shipment)
}

// ListShipmentsAdmin lists all shipments (admin only).
// @Summary      List shipments (admin)
// @Tags         Shipments
// @Produce      json
// @Security     BearerAuth
// @Param        status   query string false "Filter by legacy status"
// @Param        carrier  query string false "Filter by carrier"
// @Param        order_id query int    false "Filter by order ID"
// @Param        search   query string false "Search tracking, order #, or carrier"
// @Param        limit    query int    false "Items per page"
// @Param        offset   query int    false "Offset"
// @Success      200 {object} utils.Response{data=dto.AdminShipmentListData}
// @Router       /admin/shipments [get]
func (ctrl *ShipmentHandler) ListShipmentsAdmin(c *gin.Context) {
	var filters dto.AdminShipmentListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, 400, "invalid query parameters")
		return
	}
	filters.Limit, filters.Offset = paginationParams(c, constants.DefaultLimit)

	shipments, total, err := ctrl.shipmentService.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		RespondServiceError(c, err, "failed to list shipments")
		return
	}

	items := make([]dto.AdminShipmentListItem, 0, len(shipments))
	for i := range shipments {
		items = append(items, toAdminShipmentListItem(&shipments[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.AdminShipmentListData{
		Shipments: items,
		Total:     total,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	})
}

func toAdminShipmentListItem(s *models.Shipment) dto.AdminShipmentListItem {
	item := dto.AdminShipmentListItem{
		ID:             s.ID,
		OrderID:        s.OrderID,
		Carrier:        s.Carrier,
		TrackingNumber: s.TrackingNumber,
		Status:         s.Status,
		City:           s.City,
		Country:        s.Country,
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
	}
	if s.Order.ID != 0 {
		item.OrderNumber = s.Order.OrderNumber
	}
	if s.User.ID != 0 {
		item.CustomerName = strings.TrimSpace(s.User.FirstName + " " + s.User.LastName)
	}
	if s.EstimatedDelivery != nil {
		formatted := s.EstimatedDelivery.Format(time.RFC3339)
		item.EstimatedDelivery = &formatted
	}
	if s.ShippedAt != nil {
		formatted := s.ShippedAt.Format(time.RFC3339)
		item.ShippedAt = &formatted
	}
	if s.WorkflowState != nil {
		item.State = toStateView(s.WorkflowState)
	}
	return item
}

// GetShipment retrieves a shipment by ID (user sees own, admin sees any).
// @Summary      Get shipment by ID
// @Description  Returns a shipment. Admin can see any; users see only shipments belonging to their orders.
// @Tags         Shipments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Shipment ID"
// @Success      200  {object}  utils.Response{data=models.Shipment}
// @Failure      401  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /shipments/{id} [get]
func (ctrl *ShipmentHandler) GetShipment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	shipmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid shipment id")
		return
	}

	shipment, err := ctrl.shipmentService.GetShipmentByID(uint(shipmentID))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch shipment")
		return
	}

	// Authorisation: admin can see all, users only their own
	if !middleware.IsAdmin(c) && shipment.UserID != userID {
		utils.ForbiddenResponse(c, constants.ErrForbidden)
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, shipment)
}

type GetShipmentsByOrderRequest struct {
	OrderID uint `form:"order_id" validate:"required,gt=0"`
}

// GetShipmentsByOrder lists all shipments for a given order (user must own the order).
// @Summary      Get shipments for an order
// @Description  Returns all shipments belonging to an order (user must own the order).
// @Tags         Shipments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id  query  int  true  "Order ID"
// @Success      200       {object} utils.Response{data=[]models.Shipment}
// @Failure      400       {object} utils.Response
// @Failure      401       {object} utils.Response
// @Failure      403       {object} utils.Response
// @Failure      404       {object} utils.Response
// @Router       /shipments [get]
func (ctrl *ShipmentHandler) GetShipmentsByOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req GetShipmentsByOrderRequest
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	// Verify order ownership (reuse service or direct DB check)
	// For simplicity we use the shipment service to fetch, but we need order ownership check.
	// We'll rely on the service to filter by userID.
	// Alternatively, we can query order directly. We'll assume the service enforces ownership.
	shipments, err := ctrl.shipmentService.GetShipmentsByOrderID(req.OrderID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch shipments")
		return
	}

	// Filter by userID manually (or the service should do it)
	var userShipments []models.Shipment
	for _, s := range shipments {
		if s.UserID == userID {
			userShipments = append(userShipments, s)
		}
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, userShipments)
}

// UpdateShipmentStatus updates a shipment's status (admin only).
// Deprecated: prefer POST /workflows/shipment/{id}/transition so lifecycle hooks and audit run correctly.
// @Summary      Update shipment status (admin) [deprecated]
// @Description  Deprecated — use workflow transitions on shipment detail instead. Updates the status of a shipment and sends real-time notifications.
// @Tags         Shipments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    int     true  "Shipment ID"
// @Param        request body    object  true  "Status update request"
// @Success      200     {object} utils.Response
// @Failure      400     {object} utils.Response
// @Failure      401     {object} utils.Response
// @Failure      403     {object} utils.Response
// @Failure      404     {object} utils.Response
// @Failure      500     {object} utils.Response
// @Router       /admin/shipments/{id}/status [put]
func (ctrl *ShipmentHandler) UpdateShipmentStatus(c *gin.Context) {
	shipmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid shipment id")
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=pending processing shipped delivered"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return
	}

	if err := ctrl.shipmentService.UpdateShipmentStatus(uint(shipmentID), req.Status); err != nil {
		utils.HandleServiceError(c, err, "failed to update shipment status")
		return
	}

	utils.SuccessResponse(c, "shipment status updated successfully", nil)
}

// GetAvailableTransitions lists workflow actions allowed for a shipment (admin).
// @Summary      List shipment workflow transitions (admin)
// @Tags         Shipments
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Shipment ID"
// @Success      200 {object} utils.Response{data=dto.AvailableTransitionsView}
// @Router       /shipments/{id}/available-transitions [get]
func (ctrl *ShipmentHandler) GetAvailableTransitions(c *gin.Context) {
	shipmentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	current, transitions, err := ctrl.shipmentService.AvailableTransitions(c.Request.Context(), shipmentID)
	if err != nil {
		RespondServiceError(c, err, "failed to load shipment transitions")
		return
	}

	views := make([]dto.TransitionView, 0, len(transitions))
	for i := range transitions {
		views = append(views, toTransitionView(&transitions[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"current_state": toStateView(current),
		"transitions":   views,
	})
}

// PerformTransition applies a workflow event to a shipment (admin).
// @Summary      Transition shipment workflow state (admin)
// @Description  Fires events such as ready, pick_up, depart, out_for_delivery, deliver, or return_to_sender.
// @Tags         Shipments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Shipment ID"
// @Param        request body dto.PerformShipmentTransitionRequest true "Workflow event"
// @Success      200 {object} utils.Response{data=dto.ShipmentTransitionResponse}
// @Router       /shipments/{id}/transition [post]
func (ctrl *ShipmentHandler) PerformTransition(c *gin.Context) {
	shipmentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.PerformShipmentTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	actorRole, _ := middleware.GetUserRole(c)
	var actorIDPtr *uint
	if actorID != 0 {
		actorIDPtr = &actorID
	}

	result, err := ctrl.shipmentService.PerformTransition(
		c.Request.Context(),
		shipmentID,
		req.Event,
		req.Note,
		actorRole,
		actorIDPtr,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to transition shipment")
		return
	}

	shipment, err := ctrl.shipmentService.GetShipmentByID(shipmentID)
	if err != nil {
		RespondServiceError(c, err, "transition applied but failed to reload shipment")
		return
	}

	utils.SuccessResponse(c, "transition applied", dto.ShipmentTransitionResponse{
		Transition: toTransitionResultView(result),
		Shipment:   shipment,
	})
}

// GetShippingProviders godoc
// @Summary      Get active shipping providers
// @Description  Returns all active shipping providers (public)
// @Tags         Shipping
// @Accept       json
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.ShippingProviders}
// @Router       /shipping-providers [get]
func (ctrl *ShipmentHandler) GetShippingProviders(c *gin.Context) {
	providers, err := ctrl.shipmentService.GetShippingProviders()
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch shipping providers")
		return
	}
	utils.SuccessResponse(c, "shipping providers retrieved", providers)
}

// ListShippingProvidersAdmin godoc
// @Summary      List all shipping providers (admin)
// @Description  Returns every shipping provider including inactive ones for admin management
// @Tags         Shipping Providers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]models.ShippingProviders}
// @Failure      500 {object} utils.Response
// @Router       /admin/shipping-providers [get]
func (ctrl *ShipmentHandler) ListShippingProvidersAdmin(c *gin.Context) {
	providers, err := ctrl.shipmentService.ListShippingProvidersAdmin(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch shipping providers")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, providers)
}

// DeleteShippingProvider godoc
// @Summary      Delete a shipping provider
// @Description  Removes a shipping provider by its ID (admin only)
// @Tags         Shipping
// @Accept       json
// @Produce      json
// @Param        id   path      uint  true  "Shipping provider ID"
// @Success      200  {object}  utils.Response{data=nil}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /shipping-providers/{id} [delete]
func (ctrl *ShipmentHandler) DeleteShippingProvider(c *gin.Context) {
	// Parse the ID from the URL parameter
	idParam := c.Param("id")
	providerIdUint64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid provider id")
		return
	}

	// Convert to uint (service expects uint)
	providerId := uint(providerIdUint64)

	// Call the service
	err = ctrl.shipmentService.DeleteShippingProvider(providerId)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete shipping provider")
		return
	}

	// Success response
	utils.SuccessResponse(c, "successfully removed", nil)
}

// GetShippingProviderByID godoc
// @Summary      Get a shipping provider by ID
// @Description  Returns a single shipping provider
// @Tags         Shipping Providers
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Provider ID"
// @Success      200  {object}  utils.Response{data=models.ShippingProviders}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /shipping-providers/{id} [get]
func (ctrl *ShipmentHandler) GetShippingProviderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid provider id")
		return
	}
	provider, err := ctrl.shipmentService.GetShippingProviderByID(uint(id))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch shipping provider")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, provider)
}

// CreateShippingProvider godoc
// @Summary      Create a new shipping provider
// @Description  Adds a new shipping provider to the system
// @Tags         Shipping Providers
// @Accept       json
// @Produce      json
// @Param        request body     dto.CreateShippingProviderRequest true "Provider details"
// @Success      201     {object}  utils.Response{data=models.ShippingProviders}
// @Failure      400     {object}  utils.Response
// @Failure      409     {object}  utils.Response
// @Failure      500     {object}  utils.Response
// @Router       /shipping-providers [post]
func (ctrl *ShipmentHandler) CreateShippingProvider(c *gin.Context) {

	var req dto.CreateShippingProviderRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	provider, err := ctrl.shipmentService.CreateShippingProvider(req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create shipping provider")
		return
	}
	utils.CreatedResponse(c, constants.MsgCreateSuccess, provider)
}

// UpdateShippingProvider godoc
// @Summary      Update an existing shipping provider
// @Description  Updates fields of a shipping provider (partial update allowed)
// @Tags         Shipping Providers
// @Accept       json
// @Produce      json
// @Param        id      path      int                                 true  "Provider ID"
// @Param        request body      dto.UpdateShippingProviderRequest true "Fields to update"
// @Success      200     {object}  utils.Response{data=models.ShippingProviders}
// @Failure      400     {object}  utils.Response
// @Failure      404     {object}  utils.Response
// @Failure      500     {object}  utils.Response
// @Router       /shipping-providers/{id} [put]
func (ctrl *ShipmentHandler) UpdateShippingProvider(c *gin.Context) {
	providerId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid provider id")
		return
	}
	var req dto.UpdateShippingProviderRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	provider, err := ctrl.shipmentService.UpdateShippingProvider(uint(providerId), req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update shipping provider")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, provider)
}
