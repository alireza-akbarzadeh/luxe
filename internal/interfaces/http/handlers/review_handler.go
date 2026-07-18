package handlers

import (
	"strconv"

	appreview "github.com/alireza-akbarzadeh/luxe/internal/application/review"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReviewHandler struct {
	commands *appreview.Commands
	queries  *appreview.Queries
	validate *validator.Validate
}

func NewReviewHandler(commands *appreview.Commands, queries *appreview.Queries) *ReviewHandler {
	return &ReviewHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
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
func (rc *ReviewHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	var req dto.CreateReviewRequest
	if !utils.BindAndValidate(c, &req, rc.validate) {
		return
	}
	review, err := rc.commands.Create(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create review")
		return
	}
	review.UserID = userID
	utils.CreatedResponse(c, "review submitted for moderation", dto.EnrichReviewResponse(dto.ToReviewResponse(review, userID), review))
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
func (rc *ReviewHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateReviewRequest
	if !utils.BindAndValidate(c, &req, rc.validate) {
		return
	}
	review, err := rc.commands.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update review")
		return
	}
	utils.SuccessResponse(c, "review updated", dto.EnrichReviewResponse(dto.ToReviewResponse(review, userID), review))
}

// Delete a review
// @Summary      Delete a review
// @Description  Remove a review by ID
// @Tags         Reviews
// @Security     BearerAuth
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  utils.Response
// @Router       /reviews/{id} [delete]
func (rc *ReviewHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	err := rc.commands.Delete(c.Request.Context(), userID, id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to delete review")
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
func (rc *ReviewHandler) GetProductReviews(c *gin.Context) {
	productIDRaw, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil || productIDRaw == 0 {
		utils.ErrorResponse(c, 400, "invalid product_id")
		return
	}
	limit, offset := paginationParams(c, 10)

	reviews, total, summary, err := rc.queries.GetProductReviews(c.Request.Context(), uint(productIDRaw), limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch reviews")
		return
	}

	viewerID, _ := middleware.GetUserID(c)
	responseReviews := make([]dto.ReviewResponse, len(reviews))
	for i := range reviews {
		responseReviews[i] = dto.EnrichReviewResponse(dto.ToReviewResponse(&reviews[i], viewerID), &reviews[i])
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
func (rc *ReviewHandler) GetMyProductReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	productIDRaw, err := strconv.ParseUint(c.Query("product_id"), 10, 64)
	if err != nil || productIDRaw == 0 {
		utils.ErrorResponse(c, 400, "invalid product_id")
		return
	}

	review, err := rc.queries.GetUserReviewForProduct(c.Request.Context(), userID, uint(productIDRaw))
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch review")
		return
	}
	if review == nil {
		utils.SuccessResponse(c, constants.MsgFetchSuccess, nil)
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.EnrichReviewResponse(dto.ToReviewResponse(review, userID), review))
}

// GetMyReviews lists product reviews authored by the authenticated user.
// @Summary      List my product reviews
// @Description  Returns paginated product reviews written by the authenticated user
// @Tags         Reviews
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /users/me/reviews [get]
func (rc *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}
	limit, offset := paginationParams(c, constants.DefaultLimit)
	reviews, total, err := rc.queries.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list reviews")
		return
	}
	items := make([]dto.UserReviewResponse, len(reviews))
	for i := range reviews {
		items[i] = dto.ToUserReviewResponse(&reviews[i], userID)
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"reviews": items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// ListReviewsAdmin lists product reviews for moderation (admin only).
// @Summary      List product reviews (admin)
// @Description  Paginated list of product reviews with optional status and product filters
// @Tags         Reviews
// @Produce      json
// @Security     BearerAuth
// @Param        status     query string false "Filter by status (pending, approved, rejected)"
// @Param        product_id query int    false "Filter by product ID"
// @Param        limit      query int    false "Items per page"
// @Param        offset     query int    false "Offset"
// @Success      200 {object} utils.Response
// @Router       /admin/reviews [get]
func (rc *ReviewHandler) ListReviewsAdmin(c *gin.Context) {
	var filters dto.AdminReviewListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, 400, "invalid query parameters")
		return
	}
	filters.Limit, filters.Offset = paginationParams(c, constants.DefaultLimit)

	reviews, total, err := rc.queries.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list reviews")
		return
	}

	items := make([]dto.AdminReviewResponse, 0, len(reviews))
	for i := range reviews {
		items = append(items, dto.ToAdminReviewResponse(&reviews[i]))
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"reviews": items,
		"total":   total,
		"limit":   filters.Limit,
		"offset":  filters.Offset,
	})
}

// PerformReviewTransition applies a workflow event to a product review (admin only).
// @Summary      Transition review state (admin)
// @Description  Approve or reject a product review via workflow events (approve, reject)
// @Tags         Reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Review ID"
// @Param        request body dto.PerformReviewTransitionRequest true "Transition event"
// @Success      200 {object} utils.Response{data=dto.TransitionResultView}
// @Router       /admin/reviews/{id}/transition [post]
func (rc *ReviewHandler) PerformReviewTransition(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.PerformReviewTransitionRequest
	if !utils.BindAndValidate(c, &req, rc.validate) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	actorRole, _ := middleware.GetUserRole(c)
	var actorIDPtr *uint
	if actorID != 0 {
		actorIDPtr = &actorID
	}

	result, err := rc.commands.PerformTransition(
		c.Request.Context(),
		id,
		req.Event,
		req.Note,
		actorRole,
		actorIDPtr,
	)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to transition review")
		return
	}

	utils.SuccessResponse(c, "transition applied", toTransitionResultView(result))
}
