package handlers

import (
	"net/http"

	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentProviderHandler struct {
	service       *apppayment.Service
	walletService *appwallet.Service
	cfg           *config.Config
}

func NewPaymentMethodHandler(service *apppayment.Service, walletService *appwallet.Service, cfg *config.Config) *PaymentProviderHandler {
	return &PaymentProviderHandler{service: service, walletService: walletService, cfg: cfg}
}

// GetPaymentProviders handles GET /payment-methods
// @Summary      Get payment providers
// @Description  Returns a list of available payment providers, optionally filtering by active status.
// @Tags         Payment
// @Accept       json
// @Produce      json
// @Param        is_active query bool false "Filter by active status (default: true)"
// @Success      200 {object} utils.Response{data=[]dto.PaymentProviderResponse} "List of payment providers"
// @Failure      500 {object} utils.Response "Failed to fetch payment providers"
// @Router       /payment-providers [get]
func (h *PaymentProviderHandler) GetPaymentProviders(c *gin.Context) {
	var query dto.GetPaymentProviderQuery
	if !utils.BindAndValidateQuery(c, &query, validator.New()) {
		return
	}

	activeOnly := true
	if query.IsActive != nil {
		activeOnly = *query.IsActive
	}

	methods, err := h.service.GetPaymentProvider(c.Request.Context(), activeOnly)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch payment methods")
		return
	}

	response := make([]dto.PaymentProviderResponse, len(methods))
	for i, m := range methods {
		response[i] = dto.PaymentProviderResponse{
			Name:         m.Name,
			DisplayName:  m.DisplayName,
			Description:  m.Description,
			IconURL:      m.IconURL,
			RequiresCard: m.RequiresCard,
		}
	}

	utils.SuccessResponse(c, "Payment methods retrieved successfully", response)
}

// GetStripeConfig returns public Stripe client configuration for the frontend.
// @Summary      Get Stripe client configuration
// @Description  Returns whether Stripe is enabled and the publishable key for client-side checkout.
// @Tags         Payment
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.StripeConfigResponse}
// @Router       /payments/stripe-config [get]
func (h *PaymentProviderHandler) GetStripeConfig(c *gin.Context) {
	enabled := h.cfg != nil && h.cfg.Stripe.Enabled
	resp := dto.StripeConfigResponse{Enabled: enabled}
	if enabled {
		resp.PublishableKey = h.cfg.Stripe.PublishableKey
	}
	utils.SuccessResponse(c, "stripe configuration", resp)
}

// ListPaymentsAdmin lists order payments for the admin transactions view.
// @Summary      List payments (admin)
// @Description  Paginated, filterable list of order payments for admin transaction management.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        page      query int    false "Page number" default(1)
// @Param        limit     query int    false "Items per page" default(20)
// @Param        search    query string false "Search transaction ID, Stripe session ID, or order number"
// @Param        status    query string false "Filter by payment status"
// @Param        method    query string false "Filter by payment method"
// @Param        user_id   query int    false "Filter by user ID"
// @Param        order_id  query int    false "Filter by order ID"
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to   query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} utils.Response{data=dto.AdminPaymentListData}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /admin/payments [get]
func (h *PaymentProviderHandler) ListPaymentsAdmin(c *gin.Context) {
	var filters dto.AdminPaymentListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	filters.Page, filters.Limit = pageLimitParams(c, constants.DefaultLimit)

	payments, total, err := h.service.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list payments")
		return
	}

	items := make([]dto.AdminPaymentListItem, 0, len(payments))
	for i := range payments {
		items = append(items, dto.ToAdminPaymentListItem(&payments[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.AdminPaymentListData{
		Payments: items,
		Total:    total,
		Page:     filters.Page,
		Limit:    filters.Limit,
	})
}

// GetPaymentAdmin returns a single payment with full detail for admin review.
// @Summary      Get payment by ID (admin)
// @Description  Returns full payment details including gateway response for admin review.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Payment ID"
// @Success      200 {object} utils.Response{data=dto.AdminPaymentDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /admin/payments/{id} [get]
func (h *PaymentProviderHandler) GetPaymentAdmin(c *gin.Context) {
	paymentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	payment, err := h.service.GetByIDAdmin(c.Request.Context(), paymentID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load payment")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToAdminPaymentDetail(payment))
}

// GetPaymentsSummaryAdmin returns KPI counters for the admin payments view.
// @Summary      Get payments summary (admin)
// @Description  Returns aggregate counters (by status) and total completed volume for order payments.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.PaymentsSummaryResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /admin/payments/summary [get]
func (h *PaymentProviderHandler) GetPaymentsSummaryAdmin(c *gin.Context) {
	summary, err := h.service.GetSummaryAdmin(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load payments summary")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, summary)
}

// GetTransactionsHubSummary returns combined payment + wallet ledger KPIs for the admin transactions hub.
// @Summary      Get transactions hub summary (admin)
// @Description  Combined KPIs across order payments and wallet ledger transactions.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.TransactionsHubSummaryResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /admin/transactions/summary [get]
func (h *PaymentProviderHandler) GetTransactionsHubSummary(c *gin.Context) {
	paymentsSummary, err := h.service.GetSummaryAdmin(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load payments summary")
		return
	}
	walletSummary, err := h.walletService.GetTransactionsSummaryAdmin(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load wallet summary")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.TransactionsHubSummaryResponse{
		PaymentsTotal:   paymentsSummary.TotalCount,
		WalletTxTotal:   walletSummary.TotalCount,
		PaymentsVolume:  paymentsSummary.TotalVolume,
		WalletVolume:    walletSummary.NetVolume,
		PendingPayments: paymentsSummary.PendingCount,
		FailedPayments:  paymentsSummary.FailedCount,
	})
}
