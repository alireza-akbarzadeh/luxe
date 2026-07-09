package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	appsupport "github.com/alireza-akbarzadeh/luxe/internal/application/supportticket"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SupportHandler struct {
	commands *appsupport.Commands
	queries  *appsupport.Queries
	validate *validator.Validate
}

func NewSupportHandler(commands *appsupport.Commands, queries *appsupport.Queries) *SupportHandler {
	return &SupportHandler{commands: commands, queries: queries, validate: validator.New()}
}

// CreateSupportTicket opens a support ticket (guest or authenticated).
// @Summary      Create support ticket
// @Tags         Support
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateSupportTicketRequest true "Ticket details"
// @Success      201 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /support/tickets [post]
func (ctrl *SupportHandler) CreateSupportTicket(c *gin.Context) {
	var req dto.CreateSupportTicketRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	var userID *uint
	userEmail := ""
	userName := ""
	if uid, ok := middleware.GetUserID(c); ok && uid > 0 {
		userID = &uid
		userEmail, _ = middleware.GetUserEmail(c)
	}

	ticket, err := ctrl.commands.Create(c.Request.Context(), userID, userEmail, userName, req)
	if err != nil {
		RespondServiceError(c, err, "failed to create support ticket")
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticket.ID, 0, false)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.CreatedResponse(c, "support ticket created", resp)
}

// GetMySupportTickets lists tickets for the authenticated user.
// @Summary      List my support tickets
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page"
// @Param        offset query int false "Offset"
// @Success      200 {object} utils.Response
// @Router       /support/tickets/my [get]
func (ctrl *SupportHandler) GetMySupportTickets(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	limit, offset := paginationParams(c, constants.DefaultLimit)
	tickets, total, err := ctrl.queries.ListForUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		RespondServiceError(c, err, "failed to list support tickets")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"tickets": tickets,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetSupportTicket returns a ticket owned by the caller.
// @Summary      Get support ticket by ID
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Ticket ID"
// @Success      200 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /support/tickets/{id} [get]
func (ctrl *SupportHandler) GetSupportTicket(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticketID, userID, false)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// AddSupportTicketMessage adds a customer reply to a ticket.
// @Summary      Reply to support ticket
// @Tags         Support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Ticket ID"
// @Param        request body dto.CreateSupportTicketMessageRequest true "Message"
// @Success      201 {object} utils.Response{data=dto.SupportTicketMessageResponse}
// @Router       /support/tickets/{id}/messages [post]
func (ctrl *SupportHandler) AddSupportTicketMessage(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateSupportTicketMessageRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	uid := userID
	msg, err := ctrl.commands.AddMessage(
		c.Request.Context(),
		ticketID,
		&uid,
		constants.SupportAuthorCustomer,
		req,
		false,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to add message")
		return
	}
	utils.CreatedResponse(c, "message added", appsupport.ToMessageResponse(msg))
}

// ListSupportTicketsAdmin lists all support tickets (admin).
// @Summary      List support tickets (admin)
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Param        search   query string false "Search subject, email, or customer"
// @Param        status   query string false "Filter by status"
// @Param        channel  query string false "Filter by channel"
// @Param        priority query string false "Filter by priority"
// @Param        assignee query int    false "Filter by assignee user ID"
// @Param        limit    query int    false "Items per page"
// @Param        offset   query int    false "Offset"
// @Success      200 {object} utils.Response
// @Router       /admin/support/tickets [get]
func (ctrl *SupportHandler) ListSupportTicketsAdmin(c *gin.Context) {
	var filters dto.AdminSupportTicketListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, 400, "invalid query parameters")
		return
	}
	filters.Limit, filters.Offset = paginationParams(c, constants.DefaultLimit)

	tickets, total, err := ctrl.queries.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		RespondServiceError(c, err, "failed to list support tickets")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, gin.H{
		"tickets": tickets,
		"total":   total,
		"limit":   filters.Limit,
		"offset":  filters.Offset,
	})
}

// GetSupportTicketAdmin returns a ticket with full thread (admin).
// @Summary      Get support ticket by ID (admin)
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Ticket ID"
// @Success      200 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /admin/support/tickets/{id} [get]
func (ctrl *SupportHandler) GetSupportTicketAdmin(c *gin.Context) {
	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticketID, 0, true)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// GetSupportStatsAdmin returns aggregate support desk metrics.
// @Summary      Support desk stats (admin)
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.AdminSupportStats}
// @Router       /admin/support/stats [get]
func (ctrl *SupportHandler) GetSupportStatsAdmin(c *gin.Context) {
	stats, err := ctrl.queries.GetAdminStats(c.Request.Context())
	if err != nil {
		RespondServiceError(c, err, "failed to load support stats")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, stats)
}

// AddSupportTicketMessageAdmin adds a staff reply or internal note.
// @Summary      Add staff message (admin)
// @Tags         Support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Ticket ID"
// @Param        request body dto.CreateSupportTicketMessageRequest true "Message"
// @Success      201 {object} utils.Response{data=dto.SupportTicketMessageResponse}
// @Router       /admin/support/tickets/{id}/messages [post]
func (ctrl *SupportHandler) AddSupportTicketMessageAdmin(c *gin.Context) {
	staffID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.CreateSupportTicketMessageRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	uid := staffID
	msg, err := ctrl.commands.AddMessage(
		c.Request.Context(),
		ticketID,
		&uid,
		constants.SupportAuthorStaff,
		req,
		true,
	)
	if err != nil {
		RespondServiceError(c, err, "failed to add message")
		return
	}
	utils.CreatedResponse(c, "message added", appsupport.ToMessageResponse(msg))
}

// UpdateSupportTicketNotesAdmin updates internal admin notes.
// @Summary      Update support ticket notes (admin)
// @Tags         Support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Ticket ID"
// @Param        request body dto.UpdateSupportTicketNotesRequest true "Notes"
// @Success      200 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /admin/support/tickets/{id}/notes [patch]
func (ctrl *SupportHandler) UpdateSupportTicketNotesAdmin(c *gin.Context) {
	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSupportTicketNotesRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	if err := ctrl.commands.UpdateNotes(c.Request.Context(), ticketID, req.AdminNotes); err != nil {
		RespondServiceError(c, err, "failed to update notes")
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticketID, 0, true)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, resp)
}

// UpdateSupportTicketStatusAdmin updates ticket workflow status.
// @Summary      Update support ticket status (admin)
// @Tags         Support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Ticket ID"
// @Param        request body dto.UpdateSupportTicketStatusRequest true "Status"
// @Success      200 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /admin/support/tickets/{id}/status [patch]
func (ctrl *SupportHandler) UpdateSupportTicketStatusAdmin(c *gin.Context) {
	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSupportTicketStatusRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	if err := ctrl.commands.UpdateStatus(c.Request.Context(), ticketID, req.Status); err != nil {
		RespondServiceError(c, err, "failed to update status")
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticketID, 0, true)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, resp)
}

// UpdateSupportTicketAssigneeAdmin assigns a staff member.
// @Summary      Assign support ticket (admin)
// @Tags         Support
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Ticket ID"
// @Param        request body dto.UpdateSupportTicketAssigneeRequest true "Assignee"
// @Success      200 {object} utils.Response{data=dto.SupportTicketResponse}
// @Router       /admin/support/tickets/{id}/assign [patch]
func (ctrl *SupportHandler) UpdateSupportTicketAssigneeAdmin(c *gin.Context) {
	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateSupportTicketAssigneeRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	if err := ctrl.commands.UpdateAssignee(c.Request.Context(), ticketID, req.AssigneeID); err != nil {
		RespondServiceError(c, err, "failed to update assignee")
		return
	}

	resp, err := ctrl.queries.GetByID(c.Request.Context(), ticketID, 0, true)
	if err != nil {
		RespondServiceError(c, err, "failed to load support ticket")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, resp)
}

// SuggestSupportReplyAdmin drafts an AI reply for staff review.
// @Summary      Suggest AI reply (admin)
// @Tags         Support
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Ticket ID"
// @Success      200 {object} utils.Response{data=dto.SupportSuggestReplyResponse}
// @Router       /admin/support/tickets/{id}/suggest-reply [post]
func (ctrl *SupportHandler) SuggestSupportReplyAdmin(c *gin.Context) {
	staffID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	ticketID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	suggestion, err := ctrl.commands.SuggestReply(c.Request.Context(), staffID, ticketID)
	if err != nil {
		RespondServiceError(c, err, "failed to suggest reply")
		return
	}

	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.SupportSuggestReplyResponse{
		Suggestion: suggestion,
	})
}
