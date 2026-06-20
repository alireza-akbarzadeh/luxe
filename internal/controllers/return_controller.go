package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReturnController struct {
	svc      services.ReturnServiceInterface
	validate *validator.Validate
}

func NewReturnController(svc services.ReturnServiceInterface) *ReturnController {
	return &ReturnController{svc: svc, validate: validator.New()}
}

// CreateReturn opens a return/refund request for a delivered order.
// @Summary      Create return request
// @Tags         Returns
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateReturnRequest true "Return details"
// @Success      201 {object} utils.Response{data=dto.ReturnResponse}
// @Router       /returns [post]
func (ctrl *ReturnController) CreateReturn(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.CreateReturnRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	ret, err := ctrl.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		RespondServiceError(c, err, "failed to create return")
		return
	}
	utils.CreatedResponse(c, "return request created", toReturnResponse(ret))
}

// GetMyReturns lists the authenticated user's return requests.
// @Summary      List my returns
// @Tags         Returns
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /returns/my [get]
func (ctrl *ReturnController) GetMyReturns(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	returns, total, err := ctrl.svc.ListForUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		RespondServiceError(c, err, "failed to list returns")
		return
	}

	items := make([]dto.ReturnResponse, 0, len(returns))
	for i := range returns {
		items = append(items, toReturnResponse(&returns[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"returns": items,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetReturn returns a single return request owned by the caller.
// @Summary      Get return by ID
// @Tags         Returns
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Return ID"
// @Success      200 {object} utils.Response{data=dto.ReturnResponse}
// @Router       /returns/{id} [get]
func (ctrl *ReturnController) GetReturn(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	returnID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	ret, err := ctrl.svc.GetByID(c.Request.Context(), returnID, userID, false)
	if err != nil {
		RespondServiceError(c, err, "failed to load return")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, toReturnResponse(ret))
}

// ListReturnsAdmin lists all return requests (admin only).
// @Summary      List returns (admin)
// @Tags         Returns
// @Produce      json
// @Security     BearerAuth
// @Param        status  query string false "Filter by status"
// @Param        user_id query int    false "Filter by user ID"
// @Param        limit   query int    false "Items per page"
// @Param        offset  query int    false "Offset"
// @Success      200 {object} utils.Response
// @Router       /admin/returns [get]
func (ctrl *ReturnController) ListReturnsAdmin(c *gin.Context) {
	var filters dto.AdminReturnListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, 400, "invalid query parameters")
		return
	}
	filters.Limit, filters.Offset = paginationParams(c, constants.DefaultLimit)

	returns, total, err := ctrl.svc.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		RespondServiceError(c, err, "failed to list returns")
		return
	}

	items := make([]dto.ReturnResponse, 0, len(returns))
	for i := range returns {
		items = append(items, toReturnResponse(&returns[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"returns": items,
		"total":   total,
		"limit":   filters.Limit,
		"offset":  filters.Offset,
	})
}

// GetReturnAdmin returns a single return request (admin only).
// @Summary      Get return by ID (admin)
// @Tags         Returns
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Return ID"
// @Success      200 {object} utils.Response{data=dto.ReturnResponse}
// @Router       /admin/returns/{id} [get]
func (ctrl *ReturnController) GetReturnAdmin(c *gin.Context) {
	returnID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	ret, err := ctrl.svc.GetByID(c.Request.Context(), returnID, 0, true)
	if err != nil {
		RespondServiceError(c, err, "failed to load return")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, toReturnResponse(ret))
}

// PerformReturnTransition applies a workflow event to a return (admin only).
// @Summary      Transition return state (admin)
// @Tags         Returns
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Return ID"
// @Param        request body dto.PerformReturnTransitionRequest true "Transition event"
// @Success      200 {object} utils.Response{data=dto.TransitionResultView}
// @Router       /admin/returns/{id}/transition [post]
func (ctrl *ReturnController) PerformReturnTransition(c *gin.Context) {
	returnID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.PerformReturnTransitionRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	actorID, _ := middleware.GetUserID(c)
	actorRole, _ := middleware.GetUserRole(c)
	var actorIDPtr *uint
	if actorID != 0 {
		actorIDPtr = &actorID
	}

	result, err := ctrl.svc.PerformTransition(
		c.Request.Context(),
		returnID,
		req.Event,
		req.Note,
		actorRole,
		actorIDPtr,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to transition return")
		return
	}
	utils.SuccessResponse(c, "transition applied", toTransitionResultView(result))
}

func toReturnResponse(r *models.Return) dto.ReturnResponse {
	resp := dto.ReturnResponse{
		ID:              r.ID,
		OrderID:         r.OrderID,
		UserID:          r.UserID,
		Reason:          r.Reason,
		Status:          r.Status,
		RefundAmount:    r.RefundAmount,
		WorkflowStateID: r.WorkflowStateID,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if r.Order != nil {
		resp.OrderNumber = r.Order.OrderNumber
	}
	if r.WorkflowState != nil {
		resp.State = toStateView(r.WorkflowState)
	}
	return resp
}
