package handlers

import (
	"net/http"
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PdpHandler struct {
	pdpService     services.PdpServiceInterface
	productService services.ProductServiceInterface
	validate       *validator.Validate
}

func NewPdpHandler(pdp services.PdpServiceInterface, product services.ProductServiceInterface) *PdpHandler {
	return &PdpHandler{pdpService: pdp, productService: product, validate: validator.New()}
}

func (ctrl *PdpHandler) resolveProductID(c *gin.Context) (uint, bool) {
	identifier := c.Param("id")
	if id, err := strconv.ParseUint(identifier, 10, 64); err == nil {
		return uint(id), true
	}
	product, err := ctrl.productService.GetBySlug(identifier)
	if err != nil {
		utils.HandleServiceError(c, err, "product not found")
		return 0, false
	}
	return product.ID, true
}

func parsePdpPagination(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// GetPriceHistory godoc
// @Summary      Get product price history
// @Description  Returns price snapshots over time for PDP charts
// @Tags         Products
// @Produce      json
// @Param        id   path  string true  "Product ID or slug"
// @Param        days query int    false "Number of days to include" default(90)
// @Success      200 {object} utils.Response{data=object{points=[]dto.PriceHistoryPoint}}
// @Failure      404 {object} utils.Response
// @Router       /products/{id}/price-history [get]
func (ctrl *PdpHandler) GetPriceHistory(c *gin.Context) {
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))
	points, err := ctrl.pdpService.GetPriceHistory(productID, days)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch price history")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{"points": points})
}

// GetAlternatives godoc
// @Summary      Get cross-store product alternatives
// @Description  Lists the same product model from other stores (matched by barcode)
// @Tags         Products
// @Produce      json
// @Param        id    path  string true  "Product ID or slug"
// @Param        limit query int    false "Max items" default(6)
// @Success      200 {object} utils.Response{data=object{alternatives=[]dto.ProductAlternativeResponse}}
// @Failure      404 {object} utils.Response
// @Router       /products/{id}/alternatives [get]
func (ctrl *PdpHandler) GetAlternatives(c *gin.Context) {
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	items, err := ctrl.pdpService.GetAlternatives(c.Request.Context(), productID, limit)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch alternatives")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{"alternatives": items})
}

// GetQuestions godoc
// @Summary      List product Q&A
// @Description  Paginated buyer questions and answers for a product
// @Tags         Products
// @Produce      json
// @Param        id     path  string true  "Product ID or slug"
// @Param        limit  query int    false "Items per page" default(10)
// @Param        offset query int    false "Offset" default(0)
// @Success      200 {object} utils.Response{data=object{questions=[]dto.ProductQuestionResponse,total=int,limit=int,offset=int}}
// @Failure      404 {object} utils.Response
// @Router       /products/{id}/questions [get]
func (ctrl *PdpHandler) GetQuestions(c *gin.Context) {
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	limit, offset := parsePdpPagination(c)
	questions, total, err := ctrl.pdpService.ListQuestions(productID, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch questions")
		return
	}
	viewerID, _ := middleware.GetUserID(c)
	response := make([]dto.ProductQuestionResponse, len(questions))
	for i := range questions {
		response[i] = dto.ToProductQuestionResponse(&questions[i], viewerID)
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"questions": response,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// CreateQuestion godoc
// @Summary      Ask a product question
// @Description  Post a buyer question on the product detail page
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Product ID or slug"
// @Param        request body dto.CreateProductQuestionRequest true "Question body"
// @Success      201 {object} utils.Response{data=dto.ProductQuestionResponse}
// @Failure      401 {object} utils.Response
// @Router       /products/{id}/questions [post]
func (ctrl *PdpHandler) CreateQuestion(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	var req dto.CreateProductQuestionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	question, err := ctrl.pdpService.CreateQuestion(userID, productID, req.Body)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create question")
		return
	}
	questions, _, _ := ctrl.pdpService.ListQuestions(productID, 20, 0)
	for i := range questions {
		if questions[i].ID == question.ID {
			utils.CreatedResponse(c, "question posted", dto.ToProductQuestionResponse(&questions[i], userID))
			return
		}
	}
	utils.CreatedResponse(c, "question posted", dto.ToProductQuestionResponse(question, userID))
}

// CreateAnswer godoc
// @Summary      Answer a product question
// @Description  Post an answer; store owners are marked as store replies
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path string true "Product ID or slug"
// @Param        questionId path int    true "Question ID"
// @Param        request    body dto.CreateProductAnswerRequest true "Answer body"
// @Success      201 {object} utils.Response{data=dto.ProductAnswerResponse}
// @Failure      401 {object} utils.Response
// @Router       /products/{id}/questions/{questionId}/answers [post]
func (ctrl *PdpHandler) CreateAnswer(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid question id")
		return
	}
	var req dto.CreateProductAnswerRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	answer, err := ctrl.pdpService.CreateAnswer(userID, uint(questionID), req.Body)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create answer")
		return
	}
	utils.CreatedResponse(c, "answer posted", dto.ToProductAnswerResponse(answer))
}

// SubscribeStock godoc
// @Summary      Subscribe to back-in-stock notifications
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID or slug"
// @Success      201 {object} utils.Response{data=dto.StockNotificationStatusResponse}
// @Failure      401 {object} utils.Response
// @Router       /products/{id}/stock-notifications [post]
func (ctrl *PdpHandler) SubscribeStock(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	if err := ctrl.pdpService.SubscribeStockNotification(userID, productID); err != nil {
		utils.HandleServiceError(c, err, "failed to subscribe")
		return
	}
	utils.CreatedResponse(c, "notification subscribed", dto.StockNotificationStatusResponse{Subscribed: true})
}

// UnsubscribeStock godoc
// @Summary      Unsubscribe from back-in-stock notifications
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID or slug"
// @Success      200 {object} utils.Response{data=dto.StockNotificationStatusResponse}
// @Failure      401 {object} utils.Response
// @Router       /products/{id}/stock-notifications [delete]
func (ctrl *PdpHandler) UnsubscribeStock(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	if err := ctrl.pdpService.UnsubscribeStockNotification(userID, productID); err != nil {
		utils.HandleServiceError(c, err, "failed to unsubscribe")
		return
	}
	utils.SuccessResponse(c, "notification removed", dto.StockNotificationStatusResponse{Subscribed: false})
}

// GetStockStatus godoc
// @Summary      Get back-in-stock subscription status
// @Tags         Products
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Product ID or slug"
// @Success      200 {object} utils.Response{data=dto.StockNotificationStatusResponse}
// @Failure      401 {object} utils.Response
// @Router       /products/{id}/stock-notifications [get]
func (ctrl *PdpHandler) GetStockStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	productID, ok := ctrl.resolveProductID(c)
	if !ok {
		return
	}
	subscribed, err := ctrl.pdpService.IsStockSubscribed(userID, productID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch status")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.StockNotificationStatusResponse{Subscribed: subscribed})
}
