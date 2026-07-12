package handlers

import (
	"strconv"
	"time"

	appcheckout "github.com/alireza-akbarzadeh/luxe/internal/application/checkout"
	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	orderfacade "github.com/alireza-akbarzadeh/luxe/internal/application/order/facade"
	appstore "github.com/alireza-akbarzadeh/luxe/internal/application/store"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderHandler struct {
	orderService *orderfacade.Service
	checkoutSvc  *appcheckout.Service
	storeQueries *appstore.Queries
	validate     *validator.Validate
}

func NewOrderHandler(
	orderService *orderfacade.Service,
	checkoutSvc *appcheckout.Service,
	storeQueries *appstore.Queries,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		checkoutSvc:  checkoutSvc,
		storeQueries: storeQueries,
		validate:     validator.New(),
	}
}

// Checkout creates an order from the current cart.
// @Summary      Checkout
// @Description  Converts the authenticated user's cart into an order. For Stripe, returns checkout_url in data — redirect the customer to complete payment. Mock/wallet orders start fulfillment immediately.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CheckoutRequest true "Checkout details (address, payment info, optional shipping provider)"
// @Success      201 {object} utils.Response{data=models.Order}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /checkout [post]
func (ctrl *OrderHandler) Checkout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "invalid request body: "+err.Error())
		return
	}

	if err := ctrl.validate.Struct(req); err != nil {
		utils.ErrorResponse(c, 400, err.Error())
		return
	}

	result, err := ctrl.checkoutSvc.Checkout(c.Request.Context(), userID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to create order")
		return
	}

	utils.CreatedResponse(c, "order created successfully", result)
}

// ConfirmStripeCheckout confirms order payment after returning from Stripe Checkout.
// @Summary      Confirm Stripe checkout
// @Description  Idempotent order payment confirmation using checkout session_id from the Stripe success redirect.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ConfirmCheckoutStripeRequest true "Stripe session ID"
// @Success      200 {object} utils.Response{data=models.Order}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /checkout/confirm-stripe [post]
func (ctrl *OrderHandler) ConfirmStripeCheckout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.ConfirmCheckoutStripeRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	order, err := ctrl.checkoutSvc.ConfirmStripeOrderBySessionID(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		RespondServiceError(c, err, "failed to confirm order payment")
		return
	}

	utils.SuccessResponse(c, "payment confirmed — your order is being prepared", order)
}

// GetUserOrders returns paginated orders for the authenticated user.
// @Summary      Get user's orders
// @Description  Returns all orders for the authenticated user (paginated).
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query   int  false  "Items per page"   default(20)
// @Param        offset  query   int  false  "Offset (begin)"   default(0)
// @Success      200     {object} utils.Response{data=object{orders=[]models.Order,total=int,limit=int,offset=int}}
// @Failure      401     {object} utils.Response
// @Failure      500     {object} utils.Response
// @Router       /orders/my [get]
func (ctrl *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.OrderListFilters
	if !utils.BindAndValidateQuery(c, &req, ctrl.validate) {
		return
	}

	orders, total, err := ctrl.orderService.GetUserOrders(c.Request.Context(), userID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to get orders")
		return
	}

	data := gin.H{
		"orders": orders,
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	}
	utils.SuccessResponse(c, "orders retrieved successfully", data)
}

// ListAllOrders returns all orders with advanced filters (admin only).
// @Summary      List all orders (admin)
// @Description  Returns paginated list of all orders with filtering by status, date, amount, and user ID.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit       query   int     false  "Items per page"          default(20)
// @Param        offset      query   int     false  "Offset"                  default(0)
// @Param        status          query   string  false  "Order status"
// @Param        payment_status  query   string  false  "Payment status"
// @Param        shipment_status query   string  false  "Shipment status"
// @Param        workflow_state  query   string  false  "Order workflow state code (e.g. paid, processing, packed)"
// @Param        tag             query   string  false  "Filter by order tag"
// @Param        from_date       query   string  false  "Start date (RFC3339)"
// @Param        to_date     query   string  false  "End date (RFC3339)"
// @Param        min_amount  query   number  false  "Minimum amount"
// @Param        max_amount  query   number  false  "Maximum amount"
// @Param        user_id     query   int     false  "Filter by user ID"
// @Param        search      query   string  false  "Search order number or customer name/email"
// @Success      200         {object} utils.Response{data=dto.AdminOrderListData}
// @Failure      401         {object} utils.Response
// @Failure      403         {object} utils.Response
// @Failure      500         {object} utils.Response
// @Router       /orders [get]
func (ctrl *OrderHandler) ListAllOrders(c *gin.Context) {
	limit, offset := paginationParams(c, constants.DefaultLimit)

	// Filters
	filters := apporder.AdminOrderFilters{}
	if status := c.Query("status"); status != "" {
		filters.Status = status
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filters.FromDate = &t
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse(time.RFC3339, toDate); err == nil {
			filters.ToDate = &t
		}
	}
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if amt, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filters.MinAmount = &amt
		}
	}
	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if amt, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filters.MaxAmount = &amt
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 64); err == nil {
			filters.UserID = &[]uint{uint(id)}[0]
		}
	}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		filters.PaymentStatus = paymentStatus
	}
	if shipmentStatus := c.Query("shipment_status"); shipmentStatus != "" {
		filters.ShipmentStatus = shipmentStatus
	}
	if workflowState := c.Query("workflow_state"); workflowState != "" {
		filters.WorkflowState = workflowState
	}
	if tag := c.Query("tag"); tag != "" {
		filters.Tag = tag
	}

	orders, total, err := ctrl.orderService.GetAllOrders(c.Request.Context(), filters, limit, offset)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch orders")
		return
	}

	data := dto.AdminOrderListData{
		Orders: dto.ToAdminOrderListItems(orders),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// GetOrder returns a specific order by ID.
// @Summary      Get order by ID
// @Description  Returns a single order for the authenticated user.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Order ID"
// @Success      200  {object}  utils.Response{data=dto.AdminOrderDetailResponse}
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /orders/{id} [get]
func (ctrl *OrderHandler) GetOrder(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	role, _ := middleware.GetUserRole(c)
	if role == constants.RoleAdmin {
		order, err := ctrl.orderService.GetOrderAdmin(c.Request.Context(), orderID)
		if err != nil {
			RespondServiceError(c, err, "failed to fetch order")
			return
		}
		utils.SuccessResponse(c, "order retrieved", dto.ToAdminOrderDetail(*order))
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}
	order, err := ctrl.orderService.GetOrderByID(c.Request.Context(), orderID, userID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch order")
		return
	}
	utils.SuccessResponse(c, "order retrieved", dto.ToAdminOrderDetail(*order))
}

// UpdateOrderStatus updates an order's status (admin only).
// @Summary      Update order status (admin)
// @Description  Updates the status of an order and sends real-time notifications.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    int     true  "Order ID"
// @Param        request body    object  true  "Status update request"
// @Success      200     {object} utils.Response
// @Failure      400     {object} utils.Response
// @Failure      401     {object} utils.Response
// @Failure      403     {object} utils.Response
// @Failure      404     {object} utils.Response
// @Failure      500     {object} utils.Response
// @Router       /orders/{id}/status [put]
func (ctrl *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=pending paid shipped delivered cancelled refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, err.Error())
		return
	}

	var actorID *uint
	if uid, ok := middleware.GetUserID(c); ok {
		actorID = &uid
	}
	if err := ctrl.orderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, actorID); err != nil {
		RespondServiceError(c, err, "failed to update order status")
		return
	}

	utils.SuccessResponse(c, "order status updated successfully", nil)
}

// UpdateOrderNotes updates admin notes on an order.
// @Summary      Update order notes (admin)
// @Description  Replaces the notes field on an order.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    int                          true  "Order ID"
// @Param        request body    dto.UpdateOrderNotesRequest  true  "Notes update"
// @Success      200     {object} utils.Response{data=dto.AdminOrderDetailResponse}
// @Failure      400     {object} utils.Response
// @Failure      401     {object} utils.Response
// @Failure      403     {object} utils.Response
// @Failure      404     {object} utils.Response
// @Failure      500     {object} utils.Response
// @Router       /orders/{id}/notes [patch]
func (ctrl *OrderHandler) UpdateOrderNotes(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateOrderNotesRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	order, err := ctrl.orderService.UpdateOrderNotes(c.Request.Context(), orderID, req.Notes)
	if err != nil {
		RespondServiceError(c, err, "failed to update order notes")
		return
	}

	utils.SuccessResponse(c, "order notes updated", dto.ToAdminOrderDetail(*order))
}

// UpdateOrderTags replaces all tags on an order.
// @Summary      Update order tags (admin)
// @Description  Replaces all admin tags on an order.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path    int                         true  "Order ID"
// @Param        request body    dto.UpdateOrderTagsRequest  true  "Tags update"
// @Success      200     {object} utils.Response{data=dto.AdminOrderDetailResponse}
// @Failure      400     {object} utils.Response
// @Failure      401     {object} utils.Response
// @Failure      403     {object} utils.Response
// @Failure      404     {object} utils.Response
// @Failure      500     {object} utils.Response
// @Router       /orders/{id}/tags [put]
func (ctrl *OrderHandler) UpdateOrderTags(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateOrderTagsRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	order, err := ctrl.orderService.UpdateOrderTags(c.Request.Context(), orderID, req.Tags)
	if err != nil {
		RespondServiceError(c, err, "failed to update order tags")
		return
	}

	utils.SuccessResponse(c, "order tags updated", dto.ToAdminOrderDetail(*order))
}

// CancelOrder cancels an order belonging to the current user.
// @Summary      Cancel order
// @Description  Cancels a pending or paid order, restores stock, and refunds wallet payments.
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  int  true  "Order ID"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /orders/{id}/cancel [post]
func (ctrl *OrderHandler) CancelOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := ctrl.checkoutSvc.CancelOrder(c.Request.Context(), orderID, userID); err != nil {
		RespondServiceError(c, err, "failed to cancel order")
		return
	}
	utils.SuccessResponse(c, "order cancelled successfully", nil)
}

// GetAvailableTransitions lists workflow actions allowed for an order (admin).
// @Summary      List order workflow transitions (admin)
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Success      200 {object} utils.Response{data=dto.AvailableTransitionsView}
// @Router       /orders/{id}/available-transitions [get]
func (ctrl *OrderHandler) GetAvailableTransitions(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	current, transitions, err := ctrl.orderService.AvailableTransitions(c.Request.Context(), orderID)
	if err != nil {
		RespondServiceError(c, err, "failed to load order transitions")
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

// PerformTransition applies a workflow event to an order (admin).
// @Summary      Transition order workflow state (admin)
// @Description  Fires events such as start_processing, ship, deliver, refund, or cancel with guards and hooks.
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Order ID"
// @Param        request body dto.PerformOrderTransitionRequest true "Workflow event"
// @Success      200 {object} utils.Response{data=dto.OrderTransitionResponse}
// @Router       /orders/{id}/transition [post]
func (ctrl *OrderHandler) PerformTransition(c *gin.Context) {
	orderID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.PerformOrderTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	actorRole, _ := middleware.GetUserRole(c)
	var actorIDPtr *uint
	if actorID != 0 {
		actorIDPtr = &actorID
	}

	result, err := ctrl.orderService.PerformTransition(
		c.Request.Context(),
		orderID,
		req.Event,
		req.Note,
		actorRole,
		actorIDPtr,
		req.TrackingNumber,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to transition order")
		return
	}

	order, err := ctrl.orderService.GetOrderAdmin(c.Request.Context(), orderID)
	if err != nil {
		RespondServiceError(c, err, "transition applied but failed to reload order")
		return
	}

	utils.SuccessResponse(c, "transition applied", dto.OrderTransitionResponse{
		Transition: toTransitionResultView(result),
		Order:      order,
	})
}

func (ctrl *OrderHandler) authorizeVendorStore(c *gin.Context) (uint, uint, string, bool) {
	return authorizeVendorStore(c, ctrl.storeQueries)
}

func (ctrl *OrderHandler) parseVendorOrderFilters(c *gin.Context) apporder.AdminOrderFilters {
	filters := apporder.AdminOrderFilters{}
	if status := c.Query("status"); status != "" {
		filters.Status = status
	}
	if fromDate := c.Query("from_date"); fromDate != "" {
		if t, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filters.FromDate = &t
		}
	}
	if toDate := c.Query("to_date"); toDate != "" {
		if t, err := time.Parse(time.RFC3339, toDate); err == nil {
			filters.ToDate = &t
		}
	}
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if amt, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filters.MinAmount = &amt
		}
	}
	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if amt, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filters.MaxAmount = &amt
		}
	}
	if search := c.Query("search"); search != "" {
		filters.Search = search
	}
	return filters
}

// ListVendorStoreOrders returns paginated orders for a vendor-owned store.
// @Summary      List vendor store orders
// @Description  Returns orders containing products from the given store with filtering and search.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id          path   int     true  "Store ID"
// @Param        limit       query  int     false "Items per page" default(20)
// @Param        offset      query  int     false "Offset" default(0)
// @Param        status      query  string  false "Order status"
// @Param        from_date   query  string  false "Start date (RFC3339)"
// @Param        to_date     query  string  false "End date (RFC3339)"
// @Param        min_amount  query  number  false "Minimum amount"
// @Param        max_amount  query  number  false "Maximum amount"
// @Param        search      query  string  false "Search order number or customer"
// @Success      200 {object} utils.Response{data=dto.VendorOrderListData}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/orders [get]
func (ctrl *OrderHandler) ListVendorStoreOrders(c *gin.Context) {
	storeID, _, _, ok := ctrl.authorizeVendorStore(c)
	if !ok {
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters := ctrl.parseVendorOrderFilters(c)

	orders, total, err := ctrl.orderService.ListVendorStoreOrders(
		c.Request.Context(),
		storeID,
		filters,
		limit,
		offset,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch vendor orders")
		return
	}

	data := dto.VendorOrderListData{
		Orders: dto.ToVendorOrderListItems(orders, storeID),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// GetVendorStoreOrderStats returns order count summaries for a vendor store.
// @Summary      Vendor store order stats
// @Description  Returns total orders and counts grouped by status for the vendor dashboard.
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Store ID"
// @Success      200 {object} utils.Response{data=dto.VendorOrderStatsResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/orders/stats [get]
func (ctrl *OrderHandler) GetVendorStoreOrderStats(c *gin.Context) {
	storeID, _, _, ok := ctrl.authorizeVendorStore(c)
	if !ok {
		return
	}

	stats, err := ctrl.orderService.GetVendorStoreOrderStats(c.Request.Context(), storeID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch vendor order stats")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.VendorOrderStatsResponse{
		Total:    stats.Total,
		ByStatus: stats.ByStatus,
	})
}

// GetVendorStoreOrder returns order detail scoped to a vendor store.
// @Summary      Get vendor store order
// @Description  Returns order detail including only line items belonging to the store.
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id       path int true "Store ID"
// @Param        orderId  path int true "Order ID"
// @Success      200 {object} utils.Response{data=dto.VendorOrderDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/orders/{orderId} [get]
func (ctrl *OrderHandler) GetVendorStoreOrder(c *gin.Context) {
	storeID, _, _, ok := ctrl.authorizeVendorStore(c)
	if !ok {
		return
	}

	orderID, ok := parseUintParam(c, "orderId")
	if !ok {
		return
	}

	order, err := ctrl.orderService.GetVendorStoreOrder(c.Request.Context(), storeID, orderID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch vendor order")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToVendorOrderDetail(*order, storeID))
}

// GetVendorStoreOrderTransitions lists workflow actions a vendor may apply to an order.
// @Summary      List vendor order workflow transitions
// @Tags         Vendor
// @Produce      json
// @Security     BearerAuth
// @Param        id       path int true "Store ID"
// @Param        orderId  path int true "Order ID"
// @Success      200 {object} utils.Response{data=dto.AvailableTransitionsView}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/orders/{orderId}/available-transitions [get]
func (ctrl *OrderHandler) GetVendorStoreOrderTransitions(c *gin.Context) {
	storeID, _, _, ok := ctrl.authorizeVendorStore(c)
	if !ok {
		return
	}

	orderID, ok := parseUintParam(c, "orderId")
	if !ok {
		return
	}

	current, transitions, err := ctrl.orderService.VendorAvailableTransitions(c.Request.Context(), storeID, orderID)
	if err != nil {
		RespondServiceError(c, err, "failed to load order transitions")
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

// PerformVendorStoreOrderTransition applies a workflow event to a vendor-scoped order.
// @Summary      Transition vendor store order workflow
// @Description  Fires events such as start_processing, pack, or ship for orders containing this store's products.
// @Tags         Vendor
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path int true "Store ID"
// @Param        orderId  path int true "Order ID"
// @Param        request  body dto.PerformOrderTransitionRequest true "Workflow event"
// @Success      200 {object} utils.Response{data=dto.VendorOrderDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /vendor/stores/{id}/orders/{orderId}/transition [post]
func (ctrl *OrderHandler) PerformVendorStoreOrderTransition(c *gin.Context) {
	storeID, userID, role, ok := ctrl.authorizeVendorStore(c)
	if !ok {
		return
	}

	orderID, ok := parseUintParam(c, "orderId")
	if !ok {
		return
	}

	var req dto.PerformOrderTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	actorIDPtr := &userID
	result, err := ctrl.orderService.VendorPerformTransition(
		c.Request.Context(),
		storeID,
		orderID,
		req.Event,
		req.Note,
		role,
		actorIDPtr,
		req.TrackingNumber,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to transition order")
		return
	}

	order, err := ctrl.orderService.GetVendorStoreOrder(c.Request.Context(), storeID, orderID)
	if err != nil {
		RespondServiceError(c, err, "transition applied but failed to reload order")
		return
	}

	_ = result
	utils.SuccessResponse(c, "transition applied", dto.ToVendorOrderDetail(*order, storeID))
}
