package handlers

import (
	"net/http"

	apppayment "github.com/alireza-akbarzadeh/luxe/internal/application/payment"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentProviderHandler struct {
	service *apppayment.Service
	cfg     *config.Config
}

func NewPaymentMethodHandler(service *apppayment.Service, cfg *config.Config) *PaymentProviderHandler {
	return &PaymentProviderHandler{service: service, cfg: cfg}
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
