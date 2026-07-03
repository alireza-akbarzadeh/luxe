package handlers

import (
	"fmt"
	"net/http"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type AiHandler struct {
	aiService      *appai.Service
	searchQueries  appai.SearchQueries
	compareQueries appai.CompareQueries
}

func NewAiHandler(aiService *appai.Service, searchQueries appai.SearchQueries, compareQueries appai.CompareQueries) *AiHandler {
	return &AiHandler{aiService: aiService, searchQueries: searchQueries, compareQueries: compareQueries}
}

// GetStatus returns whether AI is enabled and which provider is configured.
// @Summary      AI status
// @Description  Returns AI provider configuration for admin UI
// @Tags         Admin AI
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AiStatusResponse}
// @Router       /admin/ai/status [get]
func (ac *AiHandler) GetStatus(c *gin.Context) {
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
func (ac *AiHandler) Generate(c *gin.Context) {
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
func (ac *AiHandler) Chat(c *gin.Context) {
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

// ProductBrief returns a structured 30-second product summary for shoppers.
// @Summary      Product AI brief
// @Description  Structured pros, cons, fit guidance, and alternatives grounded in product data
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiProductBriefRequest true "Brief request"
// @Success      200 {object} utils.Response{data=dto.AiProductBriefResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/product-brief [post]
func (ac *AiHandler) ProductBrief(c *gin.Context) {
	var req dto.AiProductBriefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.ProductBrief(c.Request.Context(), subjectKey, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == http.StatusServiceUnavailable {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
			return
		}
		utils.HandleServiceError(c, err, "ai product brief failed")
		return
	}

	utils.SuccessResponse(c, "brief", result)
}

// ShoppingAssistant runs conversational product discovery for shoppers.
// @Summary      AI shopping assistant
// @Description  Natural-language shopping assistant with follow-up questions and product recommendations
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiShoppingAssistantRequest true "Assistant request"
// @Success      200 {object} utils.Response{data=dto.AiShoppingAssistantResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/shopping-assistant [post]
func (ac *AiHandler) ShoppingAssistant(c *gin.Context) {
	var req dto.AiShoppingAssistantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.ShoppingAssistant(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == http.StatusServiceUnavailable {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
			return
		}
		utils.HandleServiceError(c, err, "ai shopping assistant failed")
		return
	}

	utils.SuccessResponse(c, "assistant", result)
}

// SearchIntent parses natural-language search into catalog filters.
// @Summary      AI search intent
// @Description  Extracts keywords, budget, and filters from a natural-language search phrase
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiSearchIntentRequest true "Search intent request"
// @Success      200 {object} utils.Response{data=dto.AiSearchIntentResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Router       /ai/search-intent [post]
func (ac *AiHandler) SearchIntent(c *gin.Context) {
	var req dto.AiSearchIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.ParseSearchIntent(c.Request.Context(), subjectKey, req.Query)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == http.StatusTooManyRequests {
			utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
			return
		}
		utils.HandleServiceError(c, err, "ai search intent failed")
		return
	}

	utils.SuccessResponse(c, "intent", result)
}

// VisualSearch finds catalog products similar to an uploaded photo.
// @Summary      AI visual search
// @Description  Analyzes a product image and returns visually similar catalog matches
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiVisualSearchRequest true "Visual search request"
// @Success      200 {object} utils.Response{data=dto.AiVisualSearchResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/visual-search [post]
func (ac *AiHandler) VisualSearch(c *gin.Context) {
	var req dto.AiVisualSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.VisualSearch(c.Request.Context(), subjectKey, ac.searchQueries, req.ImageBase64)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai visual search failed")
		return
	}

	utils.SuccessResponse(c, "visual search", result)
}

// CompareInsight returns AI explanations for side-by-side product comparison.
// @Summary      AI compare insight
// @Description  Explains trade-offs and recommendations between 2–4 compared products
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiCompareInsightRequest true "Compare insight request"
// @Success      200 {object} utils.Response{data=dto.AiCompareInsightResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/compare-insight [post]
func (ac *AiHandler) CompareInsight(c *gin.Context) {
	var req dto.AiCompareInsightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.CompareInsight(c.Request.Context(), subjectKey, ac.compareQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai compare insight failed")
		return
	}

	utils.SuccessResponse(c, "insight", result)
}
