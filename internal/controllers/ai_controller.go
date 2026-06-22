package controllers

import (
	"fmt"
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
)

type AiController struct {
	aiService services.AiServiceInterface
}

func NewAiController(aiService services.AiServiceInterface) *AiController {
	return &AiController{aiService: aiService}
}

// GetStatus returns whether AI is enabled and which provider is configured.
// @Summary      AI status
// @Description  Returns AI provider configuration for admin UI
// @Tags         Admin AI
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AiStatusResponse}
// @Router       /admin/ai/status [get]
func (ac *AiController) GetStatus(c *gin.Context) {
	utils.SuccessResponse(c, "ai status", ac.aiService.Status())
}

// Generate runs an admin AI copilot task (description, SEO, coupon copy).
// @Summary      Generate AI copy
// @Description  Generates product description, SEO metadata, or coupon copy for staff
// @Tags         Admin AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiGenerateRequest true "Generate request"
// @Success      200 {object} utils.Response{data=dto.AiGenerateResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /admin/ai/generate [post]
func (ac *AiController) Generate(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	result, err := ac.aiService.Generate(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "ai generate failed")
		return
	}

	utils.SuccessResponse(c, "generated", result)
}

// Chat answers shopper questions grounded in product data.
// @Summary      Product AI chat
// @Description  Grounded product assistant chat for shoppers (guest or authenticated)
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiChatRequest true "Chat request"
// @Success      200 {object} utils.Response{data=dto.AiChatResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/chat [post]
func (ac *AiController) Chat(c *gin.Context) {
	var req dto.AiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.Chat(c.Request.Context(), subjectKey, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == http.StatusServiceUnavailable {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
			return
		}
		utils.HandleServiceError(c, err, "ai chat failed")
		return
	}

	utils.SuccessResponse(c, "reply", result)
}
