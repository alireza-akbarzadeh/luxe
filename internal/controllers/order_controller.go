package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderController struct {
	orderService services.OrderServiceInterface
	checkoutSvc  services.CheckoutServiceInterface
	validate     *validator.Validate
}

func NewOrderController(orderService services.OrderServiceInterface, checkoutSvc services.CheckoutServiceInterface) *OrderController {
	return &OrderController{
		orderService: orderService,
		checkoutSvc:  checkoutSvc,
		validate:     validator.New(),
	}
}

// Checkout creates an order from the current cart.
// @Summary      Checkout
// @Description  Converts the authenticated user's cart into an order, creates pending payment and shipment, and starts background fulfillment.
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
func (ctrl *OrderController) Checkout(c *gin.Context) {
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
func (ctrl *OrderController) GetUserOrders(c *gin.Context) {
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
// @Param        status      query   string  false  "Order status"
// @Param        from_date   query   string  false  "Start date (RFC3339)"
// @Param        to_date     query   string  false  "End date (RFC3339)"
// @Param        min_amount  query   number  false  "Minimum amount"
// @Param        max_amount  query   number  false  "Maximum amount"
// @Param        user_id     query   int     false  "Filter by user ID"
// @Success      200         {object} utils.Response{data=object{orders=[]models.Order,total=int,limit=int,offset=int}}
// @Failure      401         {object} utils.Response
// @Failure      403         {object} utils.Response
// @Failure      500         {object} utils.Response
// @Router       /orders [get]
func (ctrl *OrderController) ListAllOrders(c *gin.Context) {
	// Pagination
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Filters
	filters := services.AdminOrderFilters{}
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

	orders, total, err := ctrl.orderService.GetAllOrders(c.Request.Context(), filters, limit, offset)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch orders")
		return
	}

	data := gin.H{
		"orders": orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
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
// @Success      200  {object}  utils.Response{data=models.Order}
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /orders/{id} [get]
func (ctrl *OrderController) GetOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := ctrl.orderService.GetOrderByID(c.Request.Context(), uint(orderID), userID)
	if err != nil {
		RespondServiceError(c, err, "failed to fetch order")
		return
	}
	utils.SuccessResponse(c, "order retrieved", order)
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
func (ctrl *OrderController) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid order id")
		return
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=pending paid shipped delivered cancelled refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, err.Error())
		return
	}

	if err := ctrl.orderService.UpdateOrderStatus(c.Request.Context(), uint(orderID), req.Status); err != nil {
		RespondServiceError(c, err, "failed to update order status")
		return
	}

	utils.SuccessResponse(c, "order status updated successfully", nil)
}
