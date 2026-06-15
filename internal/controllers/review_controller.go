package controllers

import (
	"strconv"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReviewController struct {
	reviewService services.ReviewServiceInterface
	validate      *validator.Validate
}

func NewReviewController(svc services.ReviewServiceInterface) *ReviewController {
	return &ReviewController{
		reviewService: svc,
		validate:      validator.New(),
	}
}

func parseReviewPagination(c *gin.Context) (limit, offset int) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// Create a review
// @Summary      Create product review
// @Description  Leave a rating and comment for a product
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateReviewRequest true "Review data"
// @Success      201 {object} utils.Response{data=dto.ReviewResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /reviews [post]
func (rc *ReviewController) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.CreateReviewRequest
	if !utils.BindAndValidate(c, &req, rc.validate) {
		return
	}
	review, err := rc.reviewService.Create(userID, req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to create review")
		return
	}
	review.UserID = userID
	utils.CreatedResponse(c, "review submitted", dto.ToReviewResponse(review, userID))
}

// Update a review
// @Summary      Update a review
// @Description  Modify rating or comment of an existing review
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int                       true  "Review ID"
// @Param        request body      dto.UpdateReviewRequest  true  "Updated review data"
// @Success      200     {object}  utils.Response{data=dto.ReviewResponse}
// @Router       /reviews/{id} [put]
func (rc *ReviewController) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid review id")
		return
	}
	var req dto.UpdateReviewRequest
	if !utils.BindAndValidate(c, &req, rc.validate) {
		return
	}
	review, err := rc.reviewService.Update(userID, uint(id), req)
	if err != nil {
		utils.HandleAppError(c, err, "failed to update review")
		return
	}
	utils.SuccessResponse(c, "review updated", dto.ToReviewResponse(review, userID))
}

// Delete a review
// @Summary      Delete a review
// @Description  Remove a review by ID
// @Tags         Reviews
// @Security     BearerAuth
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  utils.Response
// @Router       /reviews/{id} [delete]
func (rc *ReviewController) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid review id")
		return
	}
	err = rc.reviewService.Delete(userID, uint(id))
	if err != nil {
		utils.HandleAppError(c, err, "failed to delete review")
		return
	}
	utils.SuccessResponse(c, "review deleted", nil)
}

// GetProductReviews product reviews (public)
// @Summary      Get product reviews
// @Description  Returns paginated reviews and rating summary for a specific product
// @Tags         Reviews
// @Param        product_id query int true "Product ID"
// @Param        limit      query int false "Items per page" default(10)
// @Param        offset     query int false "Offset" default(0)
// @Success      200 {object} utils.Response
// @Router       /reviews [get]
func (rc *ReviewController) GetProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid product_id")
		return
	}
	limit, offset := parseReviewPagination(c)

	reviews, total, summary, err := rc.reviewService.GetProductReviews(uint(productID), limit, offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, err, "failed to fetch reviews")
		return
	}

	viewerID, _ := middleware.GetUserID(c)
	responseReviews := make([]dto.ReviewResponse, len(reviews))
	for i := range reviews {
		responseReviews[i] = dto.ToReviewResponse(&reviews[i], viewerID)
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"reviews": responseReviews,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"summary": summary,
	})
}

// GetMyProductReview returns the authenticated user's review for a product, if any.
// @Summary      Get my product review
// @Description  Returns the authenticated user's review for a product, if any
// @Tags         Reviews
// @Produce      json
// @Security     BearerAuth
// @Param        product_id query int true "Product ID"
// @Success      200 {object} utils.Response{data=dto.ReviewResponse}
// @Router       /reviews/me [get]
func (rc *ReviewController) GetMyProductReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	productID, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, 400, "invalid product_id")
		return
	}

	review, err := rc.reviewService.GetUserReviewForProduct(userID, uint(productID))
	if err != nil {
		utils.HandleAppError(c, err, "failed to fetch review")
		return
	}
	if review == nil {
		utils.SuccessResponse(c, constants.MsgFetchSuccess, nil)
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToReviewResponse(review, userID))
}
