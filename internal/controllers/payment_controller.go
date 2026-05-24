package controllers

import (
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PaymentProviderController struct {
	service services.PaymentServiceInterface
}

func NewPaymentMethodController(service services.PaymentServiceInterface) *PaymentProviderController {
	return &PaymentProviderController{service: service}
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
func (h *PaymentProviderController) GetPaymentProviders(c *gin.Context) {
	var query dto.GetPaymentProviderQuery
	if !utils.BindAndValidateQuery(c, &query, validator.New()) {
		return
	}

	activeOnly := true
	if query.IsActive != nil {
		activeOnly = *query.IsActive
	}

	methods, err := h.service.GetPaymentProvider(activeOnly)
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
