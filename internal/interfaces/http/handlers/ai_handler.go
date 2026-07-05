package handlers

import (
	"context"
	"fmt"
	"net/http"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	apporder "github.com/alireza-akbarzadeh/luxe/internal/application/order"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type AiHandler struct {
	aiService       *appai.Service
	searchQueries   appai.SearchQueries
	compareQueries  appai.CompareQueries
	reviewQueries   appai.ReviewSummaryQueries
	returnQueries   appai.ReturnRiskQueries
	priceHistory    appai.PriceHistoryQueries
	deliveryStats          appai.DeliveryStatsQueries
	wishlistQueries        appai.WishlistIntelligenceQueries
	shoppingMemoryQueries  appai.ShoppingMemoryQueries
	replenishmentQueries   appai.ReplenishmentQueries
}

func NewAiHandler(
	aiService *appai.Service,
	searchQueries appai.SearchQueries,
	compareQueries appai.CompareQueries,
	reviewQueries appai.ReviewSummaryQueries,
	returnQueries appai.ReturnRiskQueries,
	priceHistory appai.PriceHistoryQueries,
	deliveryStats appai.DeliveryStatsQueries,
	wishlistQueries appai.WishlistIntelligenceQueries,
	shoppingMemoryQueries appai.ShoppingMemoryQueries,
	replenishmentQueries appai.ReplenishmentQueries,
) *AiHandler {
	return &AiHandler{
		aiService:             aiService,
		searchQueries:         searchQueries,
		compareQueries:        compareQueries,
		reviewQueries:         reviewQueries,
		returnQueries:         returnQueries,
		priceHistory:          priceHistory,
		deliveryStats:         deliveryStats,
		wishlistQueries:       wishlistQueries,
		shoppingMemoryQueries: shoppingMemoryQueries,
		replenishmentQueries:  replenishmentQueries,
	}
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

// GiftFinder recommends gifts from structured recipient, occasion, and budget inputs.
// @Summary      AI gift finder
// @Description  Guided gift recommendations with follow-up questions and catalog picks
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiGiftFinderRequest true "Gift finder request"
// @Success      200 {object} utils.Response{data=dto.AiGiftFinderResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/gift-finder [post]
func (ac *AiHandler) GiftFinder(c *gin.Context) {
	var req dto.AiGiftFinderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.GiftFinder(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == http.StatusServiceUnavailable {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
			return
		}
		utils.HandleServiceError(c, err, "ai gift finder failed")
		return
	}

	utils.SuccessResponse(c, "gift finder", result)
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

// ReviewSummary synthesizes buyer reviews into scannable themes for PDP shoppers.
// @Summary      AI review summary
// @Description  Summarizes verified customer reviews into highlights and caveats
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiReviewSummaryRequest true "Review summary request"
// @Success      200 {object} utils.Response{data=dto.AiReviewSummaryResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/review-summary [post]
func (ac *AiHandler) ReviewSummary(c *gin.Context) {
	var req dto.AiReviewSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.ReviewSummary(c.Request.Context(), subjectKey, ac.reviewQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai review summary failed")
		return
	}

	utils.SuccessResponse(c, "review summary", result)
}

// ReturnRisk explains return likelihood and practical tips for PDP shoppers.
// @Summary      AI return risk insight
// @Description  Assesses return risk using product facts, reviews, and return history
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiReturnRiskRequest true "Return risk request"
// @Success      200 {object} utils.Response{data=dto.AiReturnRiskResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/return-risk [post]
func (ac *AiHandler) ReturnRisk(c *gin.Context) {
	var req dto.AiReturnRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.ReturnRisk(c.Request.Context(), subjectKey, ac.returnQueries, ac.reviewQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai return risk failed")
		return
	}

	utils.SuccessResponse(c, "return risk", result)
}

// TrustScore computes a composite trust score for PDP shoppers.
// @Summary      AI product trust score
// @Description  Scores listing trust using reviews, seller profile, listing quality, and order history
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiTrustScoreRequest true "Trust score request"
// @Success      200 {object} utils.Response{data=dto.AiTrustScoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/trust-score [post]
func (ac *AiHandler) TrustScore(c *gin.Context) {
	var req dto.AiTrustScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.TrustScore(c.Request.Context(), subjectKey, ac.returnQueries, ac.reviewQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai trust score failed")
		return
	}

	utils.SuccessResponse(c, "trust score", result)
}

// DurabilityScore estimates product longevity for PDP shoppers.
// @Summary      AI durability score
// @Description  Scores expected durability using specs, materials, and buyer review signals
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiDurabilityScoreRequest true "Durability score request"
// @Success      200 {object} utils.Response{data=dto.AiDurabilityScoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/durability-score [post]
func (ac *AiHandler) DurabilityScore(c *gin.Context) {
	var req dto.AiDurabilityScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.DurabilityScore(c.Request.Context(), subjectKey, ac.reviewQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai durability score failed")
		return
	}

	utils.SuccessResponse(c, "durability score", result)
}

// SustainabilityScore estimates environmental and ethical signals for PDP shoppers.
// @Summary      AI sustainability score
// @Description  Scores eco and ethics signals from listing specs, tags, and buyer reviews
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiSustainabilityScoreRequest true "Sustainability score request"
// @Success      200 {object} utils.Response{data=dto.AiSustainabilityScoreResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/sustainability-score [post]
func (ac *AiHandler) SustainabilityScore(c *gin.Context) {
	var req dto.AiSustainabilityScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.SustainabilityScore(c.Request.Context(), subjectKey, ac.reviewQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai sustainability score failed")
		return
	}

	utils.SuccessResponse(c, "sustainability score", result)
}

// pdpPriceHistoryAdapter bridges PDP queries into the AI price-prediction use case.
type pdpPriceHistoryAdapter struct {
	svc interface {
		GetPriceHistoryCtx(ctx context.Context, productID uint, days int) ([]dto.PriceHistoryPoint, error)
	}
}

func (a pdpPriceHistoryAdapter) GetPriceHistory(ctx context.Context, productID uint, days int) ([]dto.PriceHistoryPoint, error) {
	return a.svc.GetPriceHistoryCtx(ctx, productID, days)
}

// PricePrediction forecasts short-term price direction for PDP shoppers.
// @Summary      AI price prediction
// @Description  Analyzes recorded price history and suggests buy-now vs wait guidance
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiPricePredictionRequest true "Price prediction request"
// @Success      200 {object} utils.Response{data=dto.AiPricePredictionResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/price-prediction [post]
func (ac *AiHandler) PricePrediction(c *gin.Context) {
	var req dto.AiPricePredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.PricePrediction(c.Request.Context(), subjectKey, ac.priceHistory, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai price prediction failed")
		return
	}

	utils.SuccessResponse(c, "price prediction", result)
}

type shipmentDeliveryStatsAdapter struct {
	repo *postgres.ShipmentRepository
}

func (a shipmentDeliveryStatsAdapter) StoreDeliveryStats(ctx context.Context, storeID uint) (int64, float64, error) {
	return a.repo.StoreDeliveryStats(ctx, storeID)
}

func (a shipmentDeliveryStatsAdapter) ListActiveProviders(ctx context.Context) ([]models.ShippingProviders, error) {
	return a.repo.ListActiveProviders(ctx)
}

// DeliveryPrediction estimates delivery timing for PDP shoppers.
// @Summary      AI delivery prediction
// @Description  Estimates delivery window and speed from store shipping info and shipment history
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiDeliveryPredictionRequest true "Delivery prediction request"
// @Success      200 {object} utils.Response{data=dto.AiDeliveryPredictionResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/delivery-prediction [post]
func (ac *AiHandler) DeliveryPrediction(c *gin.Context) {
	var req dto.AiDeliveryPredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.DeliveryPrediction(c.Request.Context(), subjectKey, ac.deliveryStats, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai delivery prediction failed")
		return
	}

	utils.SuccessResponse(c, "delivery prediction", result)
}

// PurchaseAdvisor synthesizes PDP signals into a buy/wait/consider recommendation.
// @Summary      AI purchase advisor
// @Description  Recommends whether to buy now using listing, reviews, returns, and price history
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiPurchaseAdvisorRequest true "Purchase advisor request"
// @Success      200 {object} utils.Response{data=dto.AiPurchaseAdvisorResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/purchase-advisor [post]
func (ac *AiHandler) PurchaseAdvisor(c *gin.Context) {
	var req dto.AiPurchaseAdvisorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.PurchaseAdvisor(
		c.Request.Context(),
		subjectKey,
		ac.returnQueries,
		ac.reviewQueries,
		ac.priceHistory,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai purchase advisor failed")
		return
	}

	utils.SuccessResponse(c, "purchase advisor", result)
}

// SizeRecommendation suggests a size for sized products on the PDP.
// @Summary      AI size recommendation
// @Description  Recommends a size using available variants, reviews, returns, and optional shopper profile
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiSizeRecommendationRequest true "Size recommendation request"
// @Success      200 {object} utils.Response{data=dto.AiSizeRecommendationResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/size-recommendation [post]
func (ac *AiHandler) SizeRecommendation(c *gin.Context) {
	var req dto.AiSizeRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.SizeRecommendation(
		c.Request.Context(),
		subjectKey,
		ac.returnQueries,
		ac.reviewQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai size recommendation failed")
		return
	}

	utils.SuccessResponse(c, "size recommendation", result)
}

// WishlistIntelligence prioritizes items on the authenticated user's wishlist.
// @Summary      AI wishlist intelligence
// @Description  Summarizes saved items with buy now, watch, wait, and remove guidance
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiWishlistIntelligenceRequest true "Wishlist intelligence request"
// @Success      200 {object} utils.Response{data=dto.AiWishlistIntelligenceResponse}
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/wishlist-intelligence [post]
func (ac *AiHandler) WishlistIntelligence(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiWishlistIntelligenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.WishlistIntelligence(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.wishlistQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai wishlist intelligence failed")
		return
	}

	utils.SuccessResponse(c, "wishlist intelligence", result)
}

type shoppingMemoryAdapter struct {
	home     *postgres.HomeRepository
	wishlist appai.WishlistIntelligenceQueries
}

func (a shoppingMemoryAdapter) ListRecentlyViewedProductIDs(ctx context.Context, userID uint, limit int) ([]uint, error) {
	return a.home.ListRecentlyViewedProductIDs(ctx, userID, limit)
}

func (a shoppingMemoryAdapter) ListFavoriteCategoryIDs(ctx context.Context, userID uint) ([]uint, error) {
	return a.home.ListFavoriteCategoryIDs(ctx, userID)
}

func (a shoppingMemoryAdapter) GetUserWishlist(userID uint, limit, offset int, sortBy string) ([]models.Product, int64, error) {
	return a.wishlist.GetUserWishlist(userID, limit, offset, sortBy)
}

// ShoppingMemory summarizes authenticated shopper taste from recent activity.
// @Summary      AI shopping memory
// @Description  Summarizes style signals from recently viewed items, wishlist, and favorite categories
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiShoppingMemoryRequest true "Shopping memory request"
// @Success      200 {object} utils.Response{data=dto.AiShoppingMemoryResponse}
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/shopping-memory [post]
func (ac *AiHandler) ShoppingMemory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiShoppingMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.ShoppingMemory(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.shoppingMemoryQueries,
		ac.searchQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai shopping memory failed")
		return
	}

	utils.SuccessResponse(c, "shopping memory", result)
}

// GoalShopping recommends products for a stated shopping goal.
// @Summary      AI goal-based shopping
// @Description  Plans discovery steps and product picks for a shopper-defined goal
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiGoalShoppingRequest true "Goal shopping request"
// @Success      200 {object} utils.Response{data=dto.AiGoalShoppingResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/goal-shopping [post]
func (ac *AiHandler) GoalShopping(c *gin.Context) {
	var req dto.AiGoalShoppingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.GoalShopping(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai goal shopping failed")
		return
	}

	utils.SuccessResponse(c, "goal shopping", result)
}

// MoodShopping recommends products aligned with a shopper mood or vibe.
// @Summary      AI mood shopping
// @Description  Suggests style cues and catalog picks for a selected mood
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiMoodShoppingRequest true "Mood shopping request"
// @Success      200 {object} utils.Response{data=dto.AiMoodShoppingResponse}
// @Failure      400 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/mood-shopping [post]
func (ac *AiHandler) MoodShopping(c *gin.Context) {
	var req dto.AiMoodShoppingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.MoodShopping(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai mood shopping failed")
		return
	}

	utils.SuccessResponse(c, "mood shopping", result)
}

// SmartCart analyzes cart contents and suggests checkout guidance.
// @Summary      AI smart cart
// @Description  Reviews cart items and returns tips, warnings, gaps, and complementary picks
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiSmartCartRequest true "Smart cart request"
// @Success      200 {object} utils.Response{data=dto.AiSmartCartResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/smart-cart [post]
func (ac *AiHandler) SmartCart(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiSmartCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.SmartCart(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.searchQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai smart cart failed")
		return
	}

	utils.SuccessResponse(c, "smart cart", result)
}

// PersonalizedNotifications suggests notification preferences for the authenticated shopper.
// @Summary      AI personalized notifications
// @Description  Recommends order, price, wishlist, and style alerts based on shopping activity
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiPersonalizedNotificationsRequest true "Personalized notifications request"
// @Success      200 {object} utils.Response{data=dto.AiPersonalizedNotificationsResponse}
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/personalized-notifications [post]
func (ac *AiHandler) PersonalizedNotifications(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiPersonalizedNotificationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.PersonalizedNotifications(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.shoppingMemoryQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai personalized notifications failed")
		return
	}

	utils.SuccessResponse(c, "personalized notifications", result)
}

type replenishmentAdapter struct {
	orders *postgres.OrderRepository
}

func (a replenishmentAdapter) ListRecentOrders(ctx context.Context, userID uint, limit int) ([]models.Order, error) {
	if limit <= 0 {
		limit = 12
	}
	uid := userID
	orders, _, err := a.orders.List(ctx, apporder.ListFilter{
		UserID: &uid,
		Limit:  limit,
		Offset: 0,
	})
	return orders, err
}

// ReplenishmentReminders suggests reorder timing from purchase history.
// @Summary      AI replenishment reminders
// @Description  Analyzes past orders and returns likely reorder items with urgency
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiReplenishmentRemindersRequest true "Replenishment reminders request"
// @Success      200 {object} utils.Response{data=dto.AiReplenishmentRemindersResponse}
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/replenishment-reminders [post]
func (ac *AiHandler) ReplenishmentReminders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiReplenishmentRemindersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.ReplenishmentReminders(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.replenishmentQueries,
		ac.searchQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai replenishment reminders failed")
		return
	}

	utils.SuccessResponse(c, "replenishment reminders", result)
}

// HouseholdShopping recommends products for each household member profile.
// @Summary      AI household shopping
// @Description  Returns per-member catalog picks from household profiles and optional context
// @Tags         AI
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.AiHouseholdShoppingRequest true "Household shopping request"
// @Success      200 {object} utils.Response{data=dto.AiHouseholdShoppingResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/household-shopping [post]
func (ac *AiHandler) HouseholdShopping(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "unauthorized")
		return
	}

	var req dto.AiHouseholdShoppingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := fmt.Sprintf("user:%d", userID)

	result, err := ac.aiService.HouseholdShopping(
		c.Request.Context(),
		userID,
		subjectKey,
		ac.searchQueries,
		req,
	)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusUnauthorized:
				utils.UnauthorizedResponse(c, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai household shopping failed")
		return
	}

	utils.SuccessResponse(c, "household shopping", result)
}

// RoomPreview analyzes how a product fits a shopper's room photo.
// @Summary      AI room preview
// @Description  Analyzes a room photo and returns placement and styling guidance for a product
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiRoomPreviewRequest true "Room preview request"
// @Success      200 {object} utils.Response{data=dto.AiRoomPreviewResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/room-preview [post]
func (ac *AiHandler) RoomPreview(c *gin.Context) {
	var req dto.AiRoomPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.RoomPreview(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			case http.StatusNotFound:
				utils.NotFoundResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai room preview failed")
		return
	}

	utils.SuccessResponse(c, "room preview", result)
}

// VirtualTryOn analyzes how a wearable product suits a shopper photo.
// @Summary      AI virtual try-on
// @Description  Analyzes a shopper photo and returns fit and style guidance for a product
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiVirtualTryOnRequest true "Virtual try-on request"
// @Success      200 {object} utils.Response{data=dto.AiVirtualTryOnResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/virtual-try-on [post]
func (ac *AiHandler) VirtualTryOn(c *gin.Context) {
	var req dto.AiVirtualTryOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.VirtualTryOn(c.Request.Context(), subjectKey, ac.searchQueries, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			case http.StatusNotFound:
				utils.NotFoundResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai virtual try-on failed")
		return
	}

	utils.SuccessResponse(c, "virtual try-on", result)
}

// InteractiveViewer returns feature hotspots for a product image.
// @Summary      AI interactive product viewer
// @Description  Analyzes a product photo and returns clickable feature hotspots
// @Tags         AI
// @Accept       json
// @Produce      json
// @Param        body body dto.AiInteractiveViewerRequest true "Interactive viewer request"
// @Success      200 {object} utils.Response{data=dto.AiInteractiveViewerResponse}
// @Failure      400 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Failure      429 {object} utils.Response
// @Failure      503 {object} utils.Response
// @Router       /ai/interactive-viewer [post]
func (ac *AiHandler) InteractiveViewer(c *gin.Context) {
	var req dto.AiInteractiveViewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	subjectKey := c.ClientIP()
	if userID, ok := middleware.GetUserID(c); ok && userID > 0 {
		subjectKey = fmt.Sprintf("user:%d", userID)
	}

	result, err := ac.aiService.InteractiveViewer(c.Request.Context(), subjectKey, req)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case http.StatusServiceUnavailable:
				utils.ErrorResponse(c, http.StatusServiceUnavailable, appErr.Message)
				return
			case http.StatusTooManyRequests:
				utils.ErrorResponse(c, http.StatusTooManyRequests, appErr.Message)
				return
			case http.StatusBadRequest:
				utils.BadRequestResponse(c, appErr.Message)
				return
			case http.StatusNotFound:
				utils.NotFoundResponse(c, appErr.Message)
				return
			}
		}
		utils.HandleServiceError(c, err, "ai interactive viewer failed")
		return
	}

	utils.SuccessResponse(c, "interactive viewer", result)
}
